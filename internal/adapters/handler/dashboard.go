package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"html/template"
	"math"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/models"
	"github.com/ibldzn/alma/internal/utils"
)

const (
	dashboardRangeMTD    = "MTD"
	dashboardRangeYTD    = "YTD"
	dashboardRangeCustom = "custom"
)

var dashboardToday = utils.GetTodayDateInJakarta

type IndexPageData struct {
	Period               DashboardPeriod
	CurrentPositionTitle string
	Cards                DashboardCards
	Charts               DashboardCharts
	HealthTable          DashboardHealthTable
	Error                string
	CurrentUser          SessionUser
}

type DashboardPeriod struct {
	Range     string
	StartDate string
	EndDate   string
	IsMTD     bool
	IsYTD     bool
	IsCustom  bool
}

type DashboardCard struct {
	Title         string
	Value         float64
	ValueType     string
	DisplayValue  string
	Change        float64
	ChangeType    string
	DisplayChange string
	ChangeLabel   string
	ChangeTone    string
	HasData       bool
	HasChange     bool
}

type DashboardCards struct {
	TotalDeposit       DashboardCard
	ABPDeposit         DashboardCard
	NonABPDeposit      DashboardCard
	Savings            DashboardCard
	LoanFromOtherBanks DashboardCard
	ConsolidatedLDR    DashboardCard
	Items              []DashboardCard
}

type DashboardCharts struct {
	HistoricalDepositsJSON template.JS
	HistoricalLDRJSON      template.JS
	HasHistoricalDeposits  bool
	HasHistoricalLDR       bool
}

type DashboardHealthTable struct {
	Rows    []DashboardHealthTableRow
	HasRows bool
}

type DashboardHealthTableRow struct {
	No                int
	Label             string
	LDRDisplay        string
	CashRatioDisplay  string
	TotalDapemDisplay string
}

type dashboardHealthBucket struct {
	Label     string
	StartDate string
	EndDate   string
}

type lineChartData struct {
	Labels   []string           `json:"labels"`
	Datasets []lineChartDataset `json:"datasets"`
}

type lineChartDataset struct {
	Label string    `json:"label"`
	Data  []float64 `json:"data"`
}

func resolveDashboardPeriod(query url.Values, today time.Time) (DashboardPeriod, error) {
	rangeValue := strings.TrimSpace(query.Get("range"))
	if rangeValue == "" {
		rangeValue = dashboardRangeMTD
	}

	switch strings.ToUpper(rangeValue) {
	case dashboardRangeMTD:
		start := time.Date(today.Year(), today.Month(), 1, 0, 0, 0, 0, today.Location())
		return newDashboardPeriod(dashboardRangeMTD, start.Format(constants.DateFormat), today.Format(constants.DateFormat)), nil
	case dashboardRangeYTD:
		start := time.Date(today.Year(), time.January, 1, 0, 0, 0, 0, today.Location())
		return newDashboardPeriod(dashboardRangeYTD, start.Format(constants.DateFormat), today.Format(constants.DateFormat)), nil
	case "CUSTOM":
		startDate := strings.TrimSpace(query.Get("start_date"))
		endDate := strings.TrimSpace(query.Get("end_date"))
		period := newDashboardPeriod(dashboardRangeCustom, startDate, endDate)
		if _, _, err := utils.ValidateDateRange(startDate, endDate); err != nil {
			return period, err
		}
		return period, nil
	default:
		return DashboardPeriod{
			Range:     rangeValue,
			StartDate: strings.TrimSpace(query.Get("start_date")),
			EndDate:   strings.TrimSpace(query.Get("end_date")),
		}, fmt.Errorf("range must be MTD, YTD, or custom")
	}
}

func newDashboardPeriod(rangeValue, startDate, endDate string) DashboardPeriod {
	return DashboardPeriod{
		Range:     rangeValue,
		StartDate: startDate,
		EndDate:   endDate,
		IsMTD:     rangeValue == dashboardRangeMTD,
		IsYTD:     rangeValue == dashboardRangeYTD,
		IsCustom:  rangeValue == dashboardRangeCustom,
	}
}

func buildIndexPageData(
	period DashboardPeriod,
	timeDeposits []models.TimeDepositSummaryRow,
	savings []models.SavingSummaryRow,
	ldr []models.LDRSummaryRow,
	loanFromOtherBanks []models.SaldoNeraca,
) (IndexPageData, error) {
	abpDeposits, nonABPDeposits := aggregateTimeDeposits(timeDeposits)
	savingsByDate := aggregateSavings(savings)
	loanByDate := aggregateSaldoNeracas(loanFromOtherBanks)
	ldrByDate := aggregateLDR(ldr)
	totalDeposits := sumSeries(abpDeposits, nonABPDeposits, savingsByDate)

	cards := DashboardCards{
		TotalDeposit:       makeMoneyCard("Total Deposit", totalDeposits),
		ABPDeposit:         makeMoneyCard("ABP Deposit", abpDeposits),
		NonABPDeposit:      makeMoneyCard("Non ABP / Perorangan Deposit", nonABPDeposits),
		Savings:            makeMoneyCard("Savings", savingsByDate),
		LoanFromOtherBanks: makeMoneyCard("Loan from Other Banks", loanByDate),
		ConsolidatedLDR:    makeLDRCard("Consolidated LDR", ldrByDate),
	}
	cards.Items = []DashboardCard{
		cards.TotalDeposit,
		cards.ABPDeposit,
		cards.NonABPDeposit,
		cards.Savings,
		cards.LoanFromOtherBanks,
		cards.ConsolidatedLDR,
	}

	charts, err := buildDashboardCharts(abpDeposits, nonABPDeposits, savingsByDate, ldrByDate)
	if err != nil {
		return IndexPageData{}, err
	}

	return IndexPageData{
		Period:               period,
		CurrentPositionTitle: dashboardCurrentPositionTitle(period),
		Cards:                cards,
		Charts:               charts,
		HealthTable:          emptyDashboardHealthTable(),
	}, nil
}

func dashboardCurrentPositionTitle(period DashboardPeriod) string {
	endDate := strings.TrimSpace(period.EndDate)
	if endDate == "" {
		return "Current Position"
	}
	return fmt.Sprintf("Current Position: %s", formatDisplayDate(endDate))
}

func aggregateTimeDeposits(rows []models.TimeDepositSummaryRow) (map[string]float64, map[string]float64) {
	abp := make(map[string]float64)
	nonABP := make(map[string]float64)

	for _, row := range rows {
		date := strings.TrimSpace(row.Date)
		if date == "" {
			continue
		}

		switch normalizeTimeDepositProductID(row.ProductID) {
		case "202", "203":
			abp[date] += row.Balance
		case "200", "201":
			nonABP[date] += row.Balance
		}
	}

	return abp, nonABP
}

func normalizeTimeDepositProductID(productID string) string {
	productID = strings.TrimSpace(productID)
	return strings.TrimPrefix(productID, "2312")
}

func aggregateSavings(rows []models.SavingSummaryRow) map[string]float64 {
	series := make(map[string]float64)
	for _, row := range rows {
		date := strings.TrimSpace(row.Date)
		if date != "" {
			series[date] += row.Balance
		}
	}
	return series
}

func aggregateSaldoNeracas(rows []models.SaldoNeraca) map[string]float64 {
	series := make(map[string]float64)
	for _, row := range rows {
		date := strings.TrimSpace(row.Date)
		if date != "" {
			series[date] += math.Abs(row.SaldoAkhir)
		}
	}
	return series
}

func aggregateLDR(rows []models.LDRSummaryRow) map[string]float64 {
	series := make(map[string]float64)
	for _, row := range rows {
		date := strings.TrimSpace(row.Date)
		if date != "" {
			series[date] = row.ConsolidatedLDR
		}
	}
	return series
}

func aggregateCashRatio(rows []models.CashRatioSummaryRow) map[string]float64 {
	series := make(map[string]float64)
	for _, row := range rows {
		date := strings.TrimSpace(row.Date)
		if date != "" {
			series[date] = row.CashRatio
		}
	}
	return series
}

func sumSeries(seriesMaps ...map[string]float64) map[string]float64 {
	total := make(map[string]float64)
	for _, series := range seriesMaps {
		for date, value := range series {
			total[date] += value
		}
	}
	return total
}

func makeMoneyCard(title string, series map[string]float64) DashboardCard {
	return makeSeriesCard(
		title,
		"money",
		"percent",
		series,
		formatCompactRupiah,
		formatCompactRupiah,
		calculatePercentChange,
		formatSignedPercent,
		changeTone,
	)
}

func makeLDRCard(title string, series map[string]float64) DashboardCard {
	return makeSeriesCard(
		title,
		"percent",
		"percentage_point",
		series,
		formatPercent,
		formatPercentagePoint,
		calculatePercentagePointChange,
		formatSignedPercentagePoint,
		inverseChangeTone,
	)
}

func makeSeriesCard(
	title string,
	valueType string,
	changeType string,
	series map[string]float64,
	formatValue func(float64) string,
	formatDelta func(float64) string,
	calculateChange func(first, last float64) (float64, bool),
	formatChange func(float64) string,
	toneForChange func(float64) string,
) DashboardCard {
	dates := sortedMapDates(series)
	card := DashboardCard{
		Title:         title,
		ValueType:     valueType,
		ChangeType:    changeType,
		DisplayValue:  "No data",
		DisplayChange: "N/A",
		ChangeLabel:   "No data for selected period",
		ChangeTone:    "neutral",
	}

	if len(dates) == 0 {
		return card
	}

	firstDate := dates[0]
	lastDate := dates[len(dates)-1]
	first := series[firstDate]
	last := series[lastDate]
	delta := last - first
	change, ok := calculateChange(first, last)

	card.Value = last
	card.DisplayValue = formatValue(last)
	card.HasData = true
	if !ok {
		card.ChangeLabel = formatChangeLabel(delta, formatDelta(math.Abs(delta)), "N/A", firstDate, lastDate)
		card.ChangeTone = toneForChange(delta)
		return card
	}

	card.Change = change
	card.DisplayChange = formatChange(change)
	card.ChangeLabel = formatChangeLabel(delta, formatDelta(math.Abs(delta)), card.DisplayChange, firstDate, lastDate)
	card.ChangeTone = toneForChange(change)
	card.HasChange = true
	return card
}

func calculatePercentChange(first, last float64) (float64, bool) {
	if first == 0 {
		return 0, false
	}
	return (last - first) / first * 100, true
}

func calculatePercentagePointChange(first, last float64) (float64, bool) {
	return last - first, true
}

func changeTone(change float64) string {
	switch {
	case change > 0:
		return "positive"
	case change < 0:
		return "negative"
	default:
		return "neutral"
	}
}

func inverseChangeTone(change float64) string {
	switch {
	case change > 0:
		return "negative"
	case change < 0:
		return "positive"
	default:
		return "neutral"
	}
}

func formatChangeLabel(delta float64, deltaLabel, changeLabel, firstDate, lastDate string) string {
	return fmt.Sprintf(
		"%s %s (%s)\n%s \u2192 %s",
		changeDirection(delta),
		deltaLabel,
		changeLabel,
		formatDisplayDate(firstDate),
		formatDisplayDate(lastDate),
	)
}

func changeDirection(delta float64) string {
	switch {
	case delta > 0:
		return "Naik"
	case delta < 0:
		return "Turun"
	default:
		return "Tetap"
	}
}

func formatDisplayDate(date string) string {
	parsed, err := time.Parse(constants.DateFormat, date)
	if err != nil {
		return date
	}
	return parsed.Format("02 Jan 2006")
}

func (h *Handler) loadDashboardHealthTable(
	ctx context.Context,
	period DashboardPeriod,
	today time.Time,
	ldr []models.LDRSummaryRow,
	cashRatios []models.CashRatioSummaryRow,
) (DashboardHealthTable, error) {
	buckets, err := dashboardHealthBuckets(period, today)
	if err != nil {
		return DashboardHealthTable{}, err
	}

	dapemTotals := make([]models.EdapemSummaryRow, 0, len(buckets))
	for _, bucket := range buckets {
		total, err := h.EdapemService.GetTotalDapemByType(ctx, bucket.StartDate, bucket.EndDate, "1")
		if err != nil {
			return DashboardHealthTable{}, fmt.Errorf("%s: %w", bucket.Label, err)
		}
		dapemTotals = append(dapemTotals, total)
	}

	return buildDashboardHealthTable(buckets, ldr, cashRatios, dapemTotals), nil
}

func dashboardHealthBuckets(period DashboardPeriod, today time.Time) ([]dashboardHealthBucket, error) {
	start, end, err := utils.ValidateDateRange(period.StartDate, period.EndDate)
	if err != nil {
		return nil, err
	}

	var buckets []dashboardHealthBucket
	for monthStart := firstDayOfMonth(start); !monthStart.After(end); monthStart = monthStart.AddDate(0, 1, 0) {
		monthEnd := lastDayOfMonth(monthStart)
		if sameMonth(monthStart, today) {
			buckets = append(buckets, currentMonthWeekBuckets(start, end, monthStart)...)
			continue
		}

		bucketStart := maxDate(start, monthStart)
		bucketEnd := minDate(end, monthEnd)
		if bucketEnd.Before(bucketStart) {
			continue
		}

		buckets = append(buckets, dashboardHealthBucket{
			Label:     monthStart.Format("Jan 2006"),
			StartDate: bucketStart.Format(constants.DateFormat),
			EndDate:   bucketEnd.Format(constants.DateFormat),
		})
	}

	return buckets, nil
}

func currentMonthWeekBuckets(start, end, monthStart time.Time) []dashboardHealthBucket {
	monthEnd := lastDayOfMonth(monthStart)
	lastDay := monthEnd.Day()
	weekStarts := []int{1, 8, 15, 22}
	weekEnds := []int{7, 14, 21, lastDay}
	buckets := make([]dashboardHealthBucket, 0, len(weekStarts))

	for i := range weekStarts {
		weekStart := time.Date(monthStart.Year(), monthStart.Month(), weekStarts[i], 0, 0, 0, 0, monthStart.Location())
		weekEnd := time.Date(monthStart.Year(), monthStart.Month(), weekEnds[i], 0, 0, 0, 0, monthStart.Location())
		bucketStart := maxDate(start, weekStart)
		bucketEnd := minDate(end, weekEnd)
		if bucketEnd.Before(bucketStart) {
			continue
		}

		buckets = append(buckets, dashboardHealthBucket{
			Label:     fmt.Sprintf("%s W%d", monthStart.Format("Jan 2006"), i+1),
			StartDate: bucketStart.Format(constants.DateFormat),
			EndDate:   bucketEnd.Format(constants.DateFormat),
		})
	}

	return buckets
}

func firstDayOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
}

func lastDayOfMonth(date time.Time) time.Time {
	return time.Date(date.Year(), date.Month()+1, 0, 0, 0, 0, 0, date.Location())
}

func sameMonth(a, b time.Time) bool {
	return a.Year() == b.Year() && a.Month() == b.Month()
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func buildDashboardHealthTable(
	buckets []dashboardHealthBucket,
	ldr []models.LDRSummaryRow,
	cashRatios []models.CashRatioSummaryRow,
	dapemTotals []models.EdapemSummaryRow,
) DashboardHealthTable {
	ldrByDate := aggregateLDR(ldr)
	cashRatioByDate := aggregateCashRatio(cashRatios)
	rows := make([]DashboardHealthTableRow, 0, len(buckets))

	for i, bucket := range buckets {
		row := DashboardHealthTableRow{
			No:                i + 1,
			Label:             bucket.Label,
			LDRDisplay:        "No data",
			CashRatioDisplay:  "No data",
			TotalDapemDisplay: "0",
		}

		if value, ok := latestSeriesValueInBucket(ldrByDate, bucket); ok {
			row.LDRDisplay = formatPercent(value)
		}
		if value, ok := latestSeriesValueInBucket(cashRatioByDate, bucket); ok {
			row.CashRatioDisplay = formatPercent(value)
		}
		if i < len(dapemTotals) {
			row.TotalDapemDisplay = formatIntegerWithDots(float64(dapemTotals[i].TotalCustomer))
		}

		rows = append(rows, row)
	}

	return DashboardHealthTable{
		Rows:    rows,
		HasRows: len(rows) > 0,
	}
}

func latestSeriesValueInBucket(series map[string]float64, bucket dashboardHealthBucket) (float64, bool) {
	latestDate := ""
	var latestValue float64
	for date, value := range series {
		if date < bucket.StartDate || date > bucket.EndDate {
			continue
		}
		if latestDate == "" || date > latestDate {
			latestDate = date
			latestValue = value
		}
	}
	return latestValue, latestDate != ""
}

func buildDashboardCharts(
	abpDeposits map[string]float64,
	nonABPDeposits map[string]float64,
	savings map[string]float64,
	ldr map[string]float64,
) (DashboardCharts, error) {
	depositLabels := sortedSeriesDates(abpDeposits, nonABPDeposits, savings)
	ldrLabels := sortedSeriesDates(ldr)

	depositChart := lineChartData{
		Labels: depositLabels,
		Datasets: []lineChartDataset{
			{Label: "ABP", Data: valuesForLabels(abpDeposits, depositLabels)},
			{Label: "Non ABP / Perorangan", Data: valuesForLabels(nonABPDeposits, depositLabels)},
			{Label: "Savings", Data: valuesForLabels(savings, depositLabels)},
		},
	}
	ldrChart := lineChartData{
		Labels: ldrLabels,
		Datasets: []lineChartDataset{
			{Label: "Consolidated LDR", Data: valuesForLabels(ldr, ldrLabels)},
		},
	}

	depositJSON, err := marshalChartData(depositChart)
	if err != nil {
		return DashboardCharts{}, fmt.Errorf("marshal historical deposits chart: %w", err)
	}

	ldrJSON, err := marshalChartData(ldrChart)
	if err != nil {
		return DashboardCharts{}, fmt.Errorf("marshal historical LDR chart: %w", err)
	}

	return DashboardCharts{
		HistoricalDepositsJSON: depositJSON,
		HistoricalLDRJSON:      ldrJSON,
		HasHistoricalDeposits:  len(depositLabels) > 0,
		HasHistoricalLDR:       len(ldrLabels) > 0,
	}, nil
}

func marshalChartData(data lineChartData) (template.JS, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}
	return template.JS(payload), nil
}

func valuesForLabels(series map[string]float64, labels []string) []float64 {
	values := make([]float64, len(labels))
	for i, label := range labels {
		values[i] = series[label]
	}
	return values
}

func sortedSeriesDates(seriesMaps ...map[string]float64) []string {
	dateSet := make(map[string]struct{})
	for _, series := range seriesMaps {
		for date := range series {
			dateSet[date] = struct{}{}
		}
	}

	dates := make([]string, 0, len(dateSet))
	for date := range dateSet {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	return dates
}

func sortedMapDates(series map[string]float64) []string {
	dates := make([]string, 0, len(series))
	for date := range series {
		dates = append(dates, date)
	}
	sort.Strings(dates)
	return dates
}

func emptyDashboardCards() DashboardCards {
	cards := DashboardCards{
		TotalDeposit:       emptyDashboardCard("Total Deposit", "money", "percent"),
		ABPDeposit:         emptyDashboardCard("ABP Deposit", "money", "percent"),
		NonABPDeposit:      emptyDashboardCard("Non ABP / Perorangan Deposit", "money", "percent"),
		Savings:            emptyDashboardCard("Savings", "money", "percent"),
		LoanFromOtherBanks: emptyDashboardCard("Loan from Other Banks", "money", "percent"),
		ConsolidatedLDR:    emptyDashboardCard("Consolidated LDR", "percent", "percentage_point"),
	}
	cards.Items = []DashboardCard{
		cards.TotalDeposit,
		cards.ABPDeposit,
		cards.NonABPDeposit,
		cards.Savings,
		cards.LoanFromOtherBanks,
		cards.ConsolidatedLDR,
	}
	return cards
}

func emptyDashboardCard(title, valueType, changeType string) DashboardCard {
	return DashboardCard{
		Title:         title,
		ValueType:     valueType,
		ChangeType:    changeType,
		DisplayValue:  "No data",
		DisplayChange: "N/A",
		ChangeLabel:   "No data for selected period",
		ChangeTone:    "neutral",
	}
}

func emptyDashboardCharts() DashboardCharts {
	emptyJSON := template.JS(`{"labels":[],"datasets":[]}`)
	return DashboardCharts{
		HistoricalDepositsJSON: emptyJSON,
		HistoricalLDRJSON:      emptyJSON,
	}
}

func emptyDashboardHealthTable() DashboardHealthTable {
	return DashboardHealthTable{}
}

func formatCompactRupiah(value float64) string {
	sign := ""
	if value < 0 {
		sign = "-"
		value = math.Abs(value)
	}

	switch {
	case value >= 1_000_000_000_000:
		return fmt.Sprintf("%sRp %s T", sign, formatDecimalComma(value/1_000_000_000_000, 2))
	case value >= 1_000_000_000:
		return fmt.Sprintf("%sRp %s M", sign, formatDecimalComma(value/1_000_000_000, 2))
	case value >= 1_000_000:
		return fmt.Sprintf("%sRp %s jt", sign, formatDecimalComma(value/1_000_000, 2))
	default:
		return fmt.Sprintf("%sRp %s", sign, formatIntegerWithDots(math.Round(value)))
	}
}

func formatPercent(value float64) string {
	return formatDecimalComma(value, 2) + "%"
}

func formatPercentagePoint(value float64) string {
	return formatDecimalComma(value, 2) + " pp"
}

func formatSignedPercent(value float64) string {
	return formatSignedDecimal(value, 2) + "%"
}

func formatSignedPercentagePoint(value float64) string {
	return formatSignedDecimal(value, 2) + " pp"
}

func formatSignedDecimal(value float64, precision int) string {
	sign := "+"
	if value < 0 {
		sign = "-"
		value = math.Abs(value)
	}
	return sign + formatDecimalComma(value, precision)
}

func formatDecimalComma(value float64, precision int) string {
	return strings.Replace(strconv.FormatFloat(value, 'f', precision, 64), ".", ",", 1)
}

func formatIntegerWithDots(value float64) string {
	intValue := strconv.FormatInt(int64(value), 10)
	if len(intValue) <= 3 {
		return intValue
	}

	var b strings.Builder
	prefix := len(intValue) % 3
	if prefix == 0 {
		prefix = 3
	}
	b.WriteString(intValue[:prefix])
	for i := prefix; i < len(intValue); i += 3 {
		b.WriteByte('.')
		b.WriteString(intValue[i : i+3])
	}
	return b.String()
}
