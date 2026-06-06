package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"github.com/ibldzn/alma/internal/adapters/fincloud"
	"github.com/ibldzn/alma/internal/adapters/utils"
	"github.com/ibldzn/alma/models"
	"github.com/jmoiron/sqlx"
	"github.com/joho/godotenv"
)

type Database struct {
	Dwh      *sqlx.DB
	Superman *sqlx.DB
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
	defer db.Dwh.Close()
	defer db.Superman.Close()

	client, err := fincloud.NewClient(fincloud.Config{})
	ensureOk("creating Fincloud client", err)

	creds := fincloud.Credentials{
		Username: os.Getenv("FINCLOUD_USERNAME"),
		Password: os.Getenv("FINCLOUD_PASSWORD"),
	}

	session, err := client.Login(ctx, creds)
	ensureOk("logging in to Fincloud", err)
	defer client.Logout(ctx, session.ID)

	ctx = fincloud.WithFincloudSessionID(ctx, session.ID)

	report, err := client.DownloadReport(ctx, "Time Deposit Account Balance Detail Today", "")
	ensureOk("downloading report from Fincloud", err)

	headers, records, err := utils.TransposeCSV([]byte(report), '|')
	ensureOk("transposing CSV data", err)

	for _, record := range records {
		var td models.TimeDepositToday
		if err := td.FromCSV(headers, record); err != nil {
			fmt.Fprintf(os.Stderr, "parsing CSV record: %v\n", err)
			continue
		}

		fmt.Printf("%+v\n", td)
	}
}

func ensureOk(msg string, err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "%s: %v\n", msg, err)
		os.Exit(1)
	}
}

func initDB() (*Database, error) {
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
	}, nil
}
