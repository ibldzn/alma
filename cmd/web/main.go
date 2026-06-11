package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ibldzn/alma/internal/adapters/fincloud"
	"github.com/ibldzn/alma/internal/adapters/handler"
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

	supermanRepository := repositories.NewSupermanRepository(db.Superman)
	supermanService := services.NewSupermanService(supermanRepository)

	timeDepositRepository := repositories.NewTimeDepositRepository(db.AppDb, db.Dwh)
	timeDepositService := services.NewTimeDepositService(timeDepositRepository, supermanService)

	savingRepository := repositories.NewSavingRepository(db.AppDb, db.Dwh)
	savingService := services.NewSavingService(savingRepository, supermanService)

	ldrService := services.NewLDRService(supermanService)

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
		if err := fincloudService.KickOffScheduleSync(ctx); err != nil && !errors.Is(err, context.Canceled) {
			fmt.Fprintf(os.Stderr, "starting scheduled sync: %v\n", err)
			os.Exit(1)
		}
	}()

	h := handler.NewHandler(
		timeDepositService,
		savingService,
		ldrService,
		supermanService,
	)

	srv := &http.Server{
		Addr:    os.Getenv("ALMA_SERVER_ADDR"),
		Handler: h.Router(),
	}

	go func() {
		fmt.Println("starting server on " + os.Getenv("ALMA_SERVER_ADDR"))
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			fmt.Fprintf(os.Stderr, "starting HTTP server: %v\n", err)
			os.Exit(1)
		}
	}()

	<-ctx.Done()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		fmt.Fprintf(os.Stderr, "shutting down HTTP server: %v\n", err)
		os.Exit(1)
	}

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
