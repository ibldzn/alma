package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ibldzn/alma/internal/adapters/fincloud"
	"github.com/ibldzn/alma/internal/repositories"
	"github.com/ibldzn/alma/internal/services"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Database struct {
	AppDb    *sqlx.DB
	Dwh      *sqlx.DB
	Superman *sqlx.DB
}

func (db *Database) Close() {
	closeIfOk := func(name string, c *sqlx.DB) {
		if c == nil {
			return
		}
		if err := c.Close(); err != nil {
			fmt.Fprintf(os.Stderr, "closing %s database: %v\n", name, err)
		}
	}

	closeIfOk("application", db.AppDb)
	closeIfOk("DWH", db.Dwh)
	closeIfOk("Superman", db.Superman)
}

func main() {
	ensureOk("loading .env file", godotenv.Load())

	ctx, cancel := signal.NotifyContext(
		context.Background(),
		syscall.SIGINT,
		syscall.SIGTERM,
	)
	defer cancel()

	db, err := initDB()
	ensureOk("initializing database connections", err)
	defer db.Close()

	timeDepositRepository := repositories.NewTimeDepositRepository(db.Dwh, db.AppDb)
	timeDepositService := services.NewTimeDepositService(timeDepositRepository)

	savingRepository := repositories.NewSavingRepository(db.AppDb, db.Dwh)
	savingService := services.NewSavingService(savingRepository)

	fincloudService := services.NewFincloudService(
		fincloud.Config{},
		fincloud.Credentials{
			Username: os.Getenv("FINCLOUD_USERNAME"),
			Password: os.Getenv("FINCLOUD_PASSWORD"),
		},
		timeDepositService,
		savingService,
	)

	go func() {
		if err := fincloudService.KickOffScheduleSync(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "starting scheduled sync: %v\n", err)
			os.Exit(1)
		}
	}()

	x, err := timeDepositService.GetTimeDepositSummary(ctx, "2026-06-01", "2026-06-30")
	if err != nil {
		fmt.Fprintf(os.Stderr, "getting time deposit summary: %v\n", err)
		os.Exit(1)
	}

	for productID, total := range x {
		fmt.Printf("Product ID: %s, Total Nominal: %.2f\n", productID, total)
	}

	y, err := savingService.GetSavingSummary(ctx, "2026-06-01", "2026-06-30")
	if err != nil {
		fmt.Fprintf(os.Stderr, "getting saving summary: %v\n", err)
		os.Exit(1)
	}

	for productID, total := range y {
		fmt.Printf("Product ID: %s, Total Credit Balance: %.2f\n", productID, total)
	}

	<-ctx.Done()

	fmt.Println("shutting down gracefully...")
}

func ensureOk(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func initDB() (*Database, error) {
	appDb, err := sqlx.Open("mysql", os.Getenv("GOOSE_DBSTRING"))
	if err != nil {
		return nil, fmt.Errorf("opening application database: %w", err)
	}

	dwh, err := sqlx.Open("mysql", os.Getenv("DWH_DBSTRING"))
	if err != nil {
		return nil, fmt.Errorf("opening DWH database: %w", err)
	}

	superman, err := sqlx.Open("mysql", os.Getenv("SUPERMAN_DBSTRING"))
	if err != nil {
		return nil, fmt.Errorf("opening Superman database: %w", err)
	}

	return &Database{
		Dwh:      dwh,
		Superman: superman,
		AppDb:    appDb,
	}, nil
}
