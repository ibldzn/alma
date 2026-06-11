package constants

const (
	// TimeDepositHistoryTable is the name of the table in DWH that stores time deposit history.
	TimeDepositHistoryTable = "eod_time_deposit_account_balance_details"

	// TimeDepositTodayTable is the name of the table in App DB that stores today's time deposit data.
	TimeDepositTodayTable = "time_deposit_today"

	// SavingsHistoryTable is the name of the table in DWH that stores savings history.
	SavingsHistoryTable = "eod_savings_balance_details_report"

	// SavingsTodayTable is the name of the table in App DB that stores today's savings data.
	SavingsTodayTable = "savings_balance_details"

	// SupermanSaldoNeracaTable is the name of the table in SUPERMAN that stores saldo neraca data.
	SupermanSaldoNeracaTable = "saldo_neracas"

	// UsersTable is the name of the table in App DB that stores dashboard users.
	UsersTable = "users"

	// AsiaJakarta is the IANA timezone name for Asia/Jakarta.
	AsiaJakarta = "Asia/Jakarta"

	// DateFormat is the standard date format used in the application.
	DateFormat = "2006-01-02"

	// DateTimeFormat is the standard date and time format used in the application.
	DateTimeFormat = "2006-01-02 15:04:05"

	// TabInternalProductID is the product ID for Tabungan Internal, which should be excluded from summaries.
	TabInternalProductID = "TAB_INTERNAL"
)

var (
	LDRBakiDebetAccounts        = []string{"121", "122"}
	LDRFundingAccounts          = []string{"221", "2312200", "2312201"}
	LDRFundingExclusionAccounts = []string{"2212111", "2212116", "2212199"}

	CashRatioLiabilityShortTermAccounts = []string{
		"211",
		"212",
		"213",
		"219",
		"2011008",
		"2011001",
		"2011004",
		"2011005",
		"2011006",
		"2011007",
		"208",
		"221",
		"2312200",
		"2312201",
	}
	CashRatioAssetLiquidAccounts = []string{"100", "111", "112"}
)
