package constants

const (
	// TimeDepositHistoryTable is the name of the table in DWH that stores time deposit history.
	TimeDepositHistoryTable = "eod_time_deposit_account_balance_details"

	// TimeDepositTodayTable is the name of the table in App DB that stores today's time deposit data.
	TimeDepositTodayTable = "eod_time_deposit_today"

	// SavingsHistoryTable is the name of the table in DWH that stores savings history.
	SavingsHistoryTable = "eod_savings_balance_details_report"

	// SavingsTodayTable is the name of the table in App DB that stores today's savings data.
	SavingsTodayTable = "savings_today"

	// AsiaJakarta is the IANA timezone name for Asia/Jakarta.
	AsiaJakarta = "Asia/Jakarta"

	// DateFormat is the standard date format used in the application.
	DateFormat = "2006-01-02"

	// DateTimeFormat is the standard date and time format used in the application.
	DateTimeFormat = "2006-01-02 15:04:05"
)
