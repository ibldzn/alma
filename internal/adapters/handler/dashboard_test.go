package handler

import (
	"context"
	"math"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

func TestResolveDashboardPeriodDefaultsToMTD(t *testing.T) {
	today := time.Date(2026, time.June, 7, 0, 0, 0, 0, utils.JakartaLocation())

	period, err := resolveDashboardPeriod(url.Values{}, today)
	if err != nil {
		t.Fatalf("resolveDashboardPeriod returned error: %v", err)
	}

	if period.Range != dashboardRangeMTD || !period.IsMTD {
		t.Fatalf("period range = %q, IsMTD=%v; want MTD true", period.Range, period.IsMTD)
	}
	if period.StartDate != "2026-06-01" || period.EndDate != "2026-06-07" {
		t.Fatalf("period = %s to %s, want 2026-06-01 to 2026-06-07", period.StartDate, period.EndDate)
	}
}

func TestResolveDashboardPeriodYTD(t *testing.T) {
	today := time.Date(2026, time.June, 7, 0, 0, 0, 0, utils.JakartaLocation())

	period, err := resolveDashboardPeriod(url.Values{"range": {"YTD"}}, today)
	if err != nil {
		t.Fatalf("resolveDashboardPeriod returned error: %v", err)
	}

	if period.Range != dashboardRangeYTD || !period.IsYTD {
		t.Fatalf("period range = %q, IsYTD=%v; want YTD true", period.Range, period.IsYTD)
	}
	if period.StartDate != "2026-01-01" || period.EndDate != "2026-06-07" {
		t.Fatalf("period = %s to %s, want 2026-01-01 to 2026-06-07", period.StartDate, period.EndDate)
	}
}

func TestResolveDashboardPeriodCustom(t *testing.T) {
	today := time.Date(2026, time.June, 7, 0, 0, 0, 0, utils.JakartaLocation())
	query := url.Values{
		"range":      {"custom"},
		"start_date": {"2026-06-03"},
		"end_date":   {"2026-06-06"},
	}

	period, err := resolveDashboardPeriod(query, today)
	if err != nil {
		t.Fatalf("resolveDashboardPeriod returned error: %v", err)
	}

	if period.Range != dashboardRangeCustom || !period.IsCustom {
		t.Fatalf("period range = %q, IsCustom=%v; want custom true", period.Range, period.IsCustom)
	}
	if period.StartDate != "2026-06-03" || period.EndDate != "2026-06-06" {
		t.Fatalf("period = %s to %s, want 2026-06-03 to 2026-06-06", period.StartDate, period.EndDate)
	}
}

func TestResolveDashboardPeriodRejectsInvalidDate(t *testing.T) {
	today := time.Date(2026, time.June, 7, 0, 0, 0, 0, utils.JakartaLocation())
	query := url.Values{
		"range":      {"custom"},
		"start_date": {"06-03-2026"},
		"end_date":   {"2026-06-06"},
	}

	period, err := resolveDashboardPeriod(query, today)
	if err == nil {
		t.Fatal("resolveDashboardPeriod returned nil error, want invalid date error")
	}
	if period.StartDate != "06-03-2026" || period.EndDate != "2026-06-06" || !period.IsCustom {
		t.Fatalf("period did not preserve invalid custom input: %+v", period)
	}
}

func TestResolveDashboardPeriodRejectsEndBeforeStart(t *testing.T) {
	today := time.Date(2026, time.June, 7, 0, 0, 0, 0, utils.JakartaLocation())
	query := url.Values{
		"range":      {"custom"},
		"start_date": {"2026-06-06"},
		"end_date":   {"2026-06-03"},
	}

	_, err := resolveDashboardPeriod(query, today)
	if err == nil {
		t.Fatal("resolveDashboardPeriod returned nil error, want invalid range error")
	}
}

func TestBuildIndexPageDataAggregatesCardsAndCharts(t *testing.T) {
	period := newDashboardPeriod(dashboardRangeCustom, "2026-06-01", "2026-06-02")
	data, err := buildIndexPageData(
		period,
		[]models.TimeDepositSummaryRow{
			{Date: "2026-06-01", ProductID: "200", Balance: 100},
			{Date: "2026-06-01", ProductID: "201", Balance: 200},
			{Date: "2026-06-01", ProductID: "202", Balance: 300},
			{Date: "2026-06-01", ProductID: "203", Balance: 400},
			{Date: "2026-06-02", ProductID: "200", Balance: 150},
			{Date: "2026-06-02", ProductID: "202", Balance: 450},
		},
		[]models.SavingSummaryRow{
			{Date: "2026-06-01", ProductID: "101", Balance: 50},
			{Date: "2026-06-02", ProductID: "102", Balance: 70},
		},
		[]models.LDRSummaryRow{
			{Date: "2026-06-01", ConsolidatedLDR: 160},
			{Date: "2026-06-02", ConsolidatedLDR: 163.92},
		},
		[]models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "260", SaldoAkhir: -10},
			{Date: "2026-06-02", NoAkun: "260", SaldoAkhir: -15},
		},
	)
	if err != nil {
		t.Fatalf("buildIndexPageData returned error: %v", err)
	}

	assertCardValue(t, data.Cards.NonABPDeposit, 150)
	assertCardValue(t, data.Cards.ABPDeposit, 450)
	assertCardValue(t, data.Cards.Savings, 70)
	assertCardValue(t, data.Cards.TotalDeposit, 670)
	assertCardValue(t, data.Cards.LoanFromOtherBanks, 15)
	assertCardValue(t, data.Cards.ConsolidatedLDR, 163.92)

	if !data.Charts.HasHistoricalDeposits || !data.Charts.HasHistoricalLDR {
		t.Fatalf("chart data flags = deposits %v, ldr %v; want both true", data.Charts.HasHistoricalDeposits, data.Charts.HasHistoricalLDR)
	}
}

func TestMoneyChangeFromZeroIsSafe(t *testing.T) {
	card := makeMoneyCard("Savings", map[string]float64{
		"2026-06-01": 0,
		"2026-06-02": 100,
	})

	if card.HasChange {
		t.Fatal("card.HasChange = true, want false")
	}
	if card.DisplayChange != "N/A" {
		t.Fatalf("DisplayChange = %q, want N/A", card.DisplayChange)
	}
	if card.ChangeLabel != "Naik Rp 100 (N/A)\n01 Jun 2026 \u2192 02 Jun 2026" {
		t.Fatalf("ChangeLabel = %q", card.ChangeLabel)
	}
}

func TestMoneyChangeLabelFormatting(t *testing.T) {
	card := makeMoneyCard("Total Deposit", map[string]float64{
		"2026-06-01": 604_800_000_000,
		"2026-06-03": 612_360_000_000,
	})

	want := "Naik Rp 7,56 M (+1,25%)\n01 Jun 2026 \u2192 03 Jun 2026"
	if card.ChangeLabel != want {
		t.Fatalf("ChangeLabel = %q, want %q", card.ChangeLabel, want)
	}
	if card.ChangeTone != "positive" {
		t.Fatalf("ChangeTone = %q, want positive", card.ChangeTone)
	}
}

func TestLDRChangeToneIsInverted(t *testing.T) {
	decreased := makeLDRCard("Consolidated LDR", map[string]float64{
		"2026-06-01": 163.92,
		"2026-06-03": 160,
	})
	wantDecreasedLabel := "Turun 3,92 pp (-3,92 pp)\n01 Jun 2026 \u2192 03 Jun 2026"
	if decreased.ChangeLabel != wantDecreasedLabel {
		t.Fatalf("decreased ChangeLabel = %q, want %q", decreased.ChangeLabel, wantDecreasedLabel)
	}
	if decreased.ChangeTone != "positive" {
		t.Fatalf("decreased ChangeTone = %q, want positive", decreased.ChangeTone)
	}

	increased := makeLDRCard("Consolidated LDR", map[string]float64{
		"2026-06-01": 160,
		"2026-06-03": 163.92,
	})
	if increased.ChangeTone != "negative" {
		t.Fatalf("increased ChangeTone = %q, want negative", increased.ChangeTone)
	}
}

func TestFormatting(t *testing.T) {
	tests := []struct {
		name string
		got  string
		want string
	}{
		{name: "billions", got: formatCompactRupiah(612_340_000_000), want: "Rp 612,34 M"},
		{name: "trillions", got: formatCompactRupiah(1_250_000_000_000), want: "Rp 1,25 T"},
		{name: "percent", got: formatPercent(163.92), want: "163,92%"},
		{name: "percentage point", got: formatPercentagePoint(3.92), want: "3,92 pp"},
		{name: "signed percent", got: formatSignedPercent(1.25), want: "+1,25%"},
		{name: "signed pp", got: formatSignedPercentagePoint(0.24), want: "+0,24 pp"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.got != tt.want {
				t.Fatalf("got %q, want %q", tt.got, tt.want)
			}
		})
	}
}

func TestIndexRendersHTMLDashboard(t *testing.T) {
	timeDepositService := &fakeTimeDepositService{
		summaryRows: []models.TimeDepositSummaryRow{
			{Date: "2026-06-01", ProductID: "200", Balance: 100_000_000_000},
			{Date: "2026-06-01", ProductID: "202", Balance: 200_000_000_000},
			{Date: "2026-06-02", ProductID: "200", Balance: 120_000_000_000},
			{Date: "2026-06-02", ProductID: "202", Balance: 220_000_000_000},
		},
	}
	savingService := &fakeSavingService{
		summaryRows: []models.SavingSummaryRow{
			{Date: "2026-06-01", ProductID: "101", Balance: 50_000_000_000},
			{Date: "2026-06-02", ProductID: "101", Balance: 60_000_000_000},
		},
	}
	ldrService := &fakeLDRService{
		rows: []models.LDRSummaryRow{
			{Date: "2026-06-01", ConsolidatedLDR: 160},
			{Date: "2026-06-02", ConsolidatedLDR: 163.92},
		},
	}
	supermanService := &fakeSupermanService{
		rows: []models.SaldoNeraca{
			{Date: "2026-06-01", NoAkun: "260", SaldoAkhir: -10_000_000_000},
			{Date: "2026-06-02", NoAkun: "260", SaldoAkhir: -15_000_000_000},
		},
	}
	handler := newTestHandler(t, timeDepositService, savingService, ldrService, supermanService)

	req := httptest.NewRequest(http.MethodGet, "/?range=custom&start_date=2026-06-01&end_date=2026-06-02", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if contentType := rec.Header().Get("Content-Type"); contentType != "text/html; charset=utf-8" {
		t.Fatalf("Content-Type = %q, want text/html; charset=utf-8", contentType)
	}

	body := rec.Body.String()
	for _, want := range []string{
		"Funding &amp; Liquidity",
		"Total Deposit",
		"Rp 400,00 M",
		"historical-deposits-data",
		"/assets/js/chart.umd.min.js",
		"filter-menu-toggle",
		`aria-controls="dashboard-filters"`,
		`id="dashboard-filters"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("response body missing %q", want)
		}
	}

	if timeDepositService.startDate != "2026-06-01" || timeDepositService.endDate != "2026-06-02" {
		t.Fatalf("time deposit service date range = %s to %s", timeDepositService.startDate, timeDepositService.endDate)
	}
	if !reflect.DeepEqual(supermanService.accounts, []string{"260"}) {
		t.Fatalf("superman accounts = %v, want [260]", supermanService.accounts)
	}
}

func TestIndexDateInputsDisabledOutsideCustomRange(t *testing.T) {
	tests := []struct {
		name         string
		target       string
		wantDisabled bool
	}{
		{name: "MTD", target: "/?range=MTD", wantDisabled: true},
		{name: "YTD", target: "/?range=YTD", wantDisabled: true},
		{name: "custom", target: "/?range=custom&start_date=2026-06-01&end_date=2026-06-02", wantDisabled: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := newTestHandler(t, nil, nil, nil, nil)
			req := httptest.NewRequest(http.MethodGet, tt.target, nil)
			rec := httptest.NewRecorder()

			handler.Router().ServeHTTP(rec, req)

			if rec.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
			}

			body := rec.Body.String()
			for _, name := range []string{"start_date", "end_date"} {
				if got := inputTagHasAttribute(body, name, "disabled"); got != tt.wantDisabled {
					t.Fatalf("%s disabled = %v, want %v", name, got, tt.wantDisabled)
				}
			}
		})
	}
}

func TestIndexInvalidQueryRendersHTMLWithoutServiceCalls(t *testing.T) {
	timeDepositService := &fakeTimeDepositService{}
	handler := newTestHandler(t, timeDepositService, nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/?range=custom&start_date=bad&end_date=2026-06-02", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if timeDepositService.called {
		t.Fatal("time deposit service was called for invalid query")
	}
	if body := rec.Body.String(); !strings.Contains(body, "Invalid date filter") || !strings.Contains(body, `value="bad"`) {
		t.Fatalf("response body did not render invalid filter state: %s", body)
	}
}

func TestDashboardAllowsAnonymousAccess(t *testing.T) {
	timeDepositService := &fakeTimeDepositService{}
	handler := newTestHandler(t, timeDepositService, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if !timeDepositService.called {
		t.Fatal("time deposit service was not called")
	}
}

func TestAssetsStayPublic(t *testing.T) {
	handler := newTestHandler(t, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/assets/css/dashboard.css", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if body := rec.Body.String(); !strings.Contains(body, ":root") {
		t.Fatalf("asset body missing css root: %s", body)
	}
}

func TestLoginFormRedirectsToDashboard(t *testing.T) {
	handler := newTestHandler(t, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodGet, "/login?next=%2F", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
}

func TestLoginSubmitRedirectsToDashboardWithoutCookie(t *testing.T) {
	handler := newTestHandler(t, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("username=wrong&password=bad&next=%2F"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
	if cookies := rec.Result().Cookies(); len(cookies) != 0 {
		t.Fatalf("cookies = %+v, want none", cookies)
	}
}

func TestLogoutClearsSessionCookie(t *testing.T) {
	handler := newTestHandler(t, nil, nil, nil, nil)
	req := httptest.NewRequest(http.MethodPost, "/logout", nil)
	rec := httptest.NewRecorder()

	handler.Router().ServeHTTP(rec, req)

	if rec.Code != http.StatusSeeOther {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusSeeOther)
	}
	if location := rec.Header().Get("Location"); location != "/" {
		t.Fatalf("Location = %q, want /", location)
	}
	cookies := rec.Result().Cookies()
	if len(cookies) != 1 || cookies[0].Name != sessionCookieName || cookies[0].MaxAge >= 0 {
		t.Fatalf("clear cookie = %+v", cookies)
	}
}

func inputTagHasAttribute(body, name, attr string) bool {
	nameIndex := strings.Index(body, `name="`+name+`"`)
	if nameIndex == -1 {
		return false
	}

	tagEndOffset := strings.Index(body[nameIndex:], ">")
	if tagEndOffset == -1 {
		return false
	}

	return strings.Contains(body[nameIndex:nameIndex+tagEndOffset], attr)
}

func assertCardValue(t *testing.T, card DashboardCard, want float64) {
	t.Helper()

	if math.Abs(card.Value-want) > 0.0001 {
		t.Fatalf("%s value = %v, want %v", card.Title, card.Value, want)
	}
}

func newTestHandler(
	t *testing.T,
	timeDepositService *fakeTimeDepositService,
	savingService *fakeSavingService,
	ldrService *fakeLDRService,
	supermanService *fakeSupermanService,
) *Handler {
	t.Helper()

	if timeDepositService == nil {
		timeDepositService = &fakeTimeDepositService{}
	}
	if savingService == nil {
		savingService = &fakeSavingService{}
	}
	if ldrService == nil {
		ldrService = &fakeLDRService{}
	}
	if supermanService == nil {
		supermanService = &fakeSupermanService{}
	}
	return NewHandler(
		timeDepositService,
		savingService,
		ldrService,
		supermanService,
	)
}

type fakeTimeDepositService struct {
	summaryRows []models.TimeDepositSummaryRow
	err         error
	called      bool
	startDate   string
	endDate     string
}

func (f *fakeTimeDepositService) GetTimeDepositHistory(ctx context.Context, startDate, endDate string) ([]models.TimeDeposit, error) {
	return nil, nil
}

func (f *fakeTimeDepositService) UpsertTimeDeposits(ctx context.Context, timeDeposits []models.TimeDeposit) error {
	return nil
}

func (f *fakeTimeDepositService) GetTimeDepositSummary(ctx context.Context, startDate, endDate string) ([]models.TimeDepositSummaryRow, error) {
	f.called = true
	f.startDate = startDate
	f.endDate = endDate
	return f.summaryRows, f.err
}

type fakeSavingService struct {
	summaryRows []models.SavingSummaryRow
	err         error
}

func (f *fakeSavingService) GetSavingHistory(ctx context.Context, date string) ([]models.Saving, error) {
	return nil, nil
}

func (f *fakeSavingService) UpsertSavings(ctx context.Context, savings []models.Saving) error {
	return nil
}

func (f *fakeSavingService) GetSavingSummary(ctx context.Context, startDate, endDate string) ([]models.SavingSummaryRow, error) {
	return f.summaryRows, f.err
}

type fakeLDRService struct {
	rows []models.LDRSummaryRow
	err  error
}

func (f *fakeLDRService) GetLDRHistory(ctx context.Context, startDate, endDate string) ([]models.LDRSummaryRow, error) {
	return f.rows, f.err
}

type fakeSupermanService struct {
	rows     []models.SaldoNeraca
	err      error
	accounts []string
}

func (f *fakeSupermanService) GetSaldoNeracas(ctx context.Context, startDate, endDate string, accounts []string) ([]models.SaldoNeraca, error) {
	f.accounts = append([]string(nil), accounts...)
	return f.rows, f.err
}
