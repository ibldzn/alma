package handler

import (
	"bytes"
	"html/template"
	"io/fs"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/ibldzn/alma/internal/interfaces"
	"github.com/ibldzn/alma/internal/web"
)

type Handler struct {
	TimeDepositService interfaces.ITimeDepositService
	SavingService      interfaces.ISavingService
	TKSService         interfaces.ITKSService
	SupermanService    interfaces.ISupermanService
	templates          *template.Template
	assetsHandler      http.Handler
	EdapemService      interfaces.IEdapemService
}

func NewHandler(
	timeDepositService interfaces.ITimeDepositService,
	savingService interfaces.ISavingService,
	tksService interfaces.ITKSService,
	supermanService interfaces.ISupermanService,
	edapemService interfaces.IEdapemService,
) *Handler {
	templates := template.Must(template.ParseFS(web.Files, "templates/*.html"))

	staticFS, err := fs.Sub(web.Files, "static")
	if err != nil {
		panic(err)
	}

	return &Handler{
		TimeDepositService: timeDepositService,
		SavingService:      savingService,
		TKSService:         tksService,
		SupermanService:    supermanService,
		EdapemService:      edapemService,
		templates:          templates,
		assetsHandler:      http.StripPrefix("/assets/", http.FileServer(http.FS(staticFS))),
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Handle("/assets/*", h.assetsHandler)
	r.Get("/login", h.LoginForm)
	r.Post("/login", h.LoginSubmit)
	r.Post("/logout", h.Logout)
	r.Get("/", h.Index)

	return r
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	var currentUser SessionUser
	today := dashboardToday()
	period, err := resolveDashboardPeriod(r.URL.Query(), today)
	if err != nil {
		h.renderIndex(w, http.StatusBadRequest, IndexPageData{
			Period:               period,
			CurrentPositionTitle: dashboardCurrentPositionTitle(period),
			Cards:                emptyDashboardCards(),
			Charts:               emptyDashboardCharts(),
			HealthTable:          emptyDashboardHealthTable(),
			Error:                "Invalid date filter: " + err.Error(),
			CurrentUser:          currentUser,
		})
		return
	}

	timeDeposits, err := h.TimeDepositService.GetTimeDepositSummary(r.Context(), period.StartDate, period.EndDate)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load time deposit data: "+err.Error())
		return
	}

	savings, err := h.SavingService.GetSavingSummary(r.Context(), period.StartDate, period.EndDate)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load savings data: "+err.Error())
		return
	}

	ldr, err := h.TKSService.GetLDRHistory(r.Context(), period.StartDate, period.EndDate)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load LDR data: "+err.Error())
		return
	}

	cashRatios, err := h.TKSService.GetCashRatioHistory(r.Context(), period.StartDate, period.EndDate)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load Cash Ratio data: "+err.Error())
		return
	}

	loanFromOtherBanks, err := h.SupermanService.GetSaldoNeracas(r.Context(), period.StartDate, period.EndDate, []string{"260"})
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load Pinjaman Bank Lain data: "+err.Error())
		return
	}

	healthTable, err := h.loadDashboardHealthTable(r.Context(), period, today, ldr, cashRatios)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to load Dapem data: "+err.Error())
		return
	}

	data, err := buildIndexPageData(period, timeDeposits, savings, ldr, loanFromOtherBanks)
	if err != nil {
		h.renderDashboardLoadError(w, period, currentUser, "Unable to prepare dashboard data: "+err.Error())
		return
	}
	data.CurrentUser = currentUser
	data.HealthTable = healthTable

	h.renderIndex(w, http.StatusOK, data)
}

func (h *Handler) renderDashboardLoadError(w http.ResponseWriter, period DashboardPeriod, currentUser SessionUser, message string) {
	h.renderIndex(w, http.StatusInternalServerError, IndexPageData{
		Period:               period,
		CurrentPositionTitle: dashboardCurrentPositionTitle(period),
		Cards:                emptyDashboardCards(),
		Charts:               emptyDashboardCharts(),
		HealthTable:          emptyDashboardHealthTable(),
		Error:                message,
		CurrentUser:          currentUser,
	})
}

func (h *Handler) renderIndex(w http.ResponseWriter, status int, data IndexPageData) {
	h.renderTemplate(w, status, "index.html", data)
}

func (h *Handler) renderTemplate(w http.ResponseWriter, status int, name string, data any) {
	var body bytes.Buffer
	if err := h.templates.ExecuteTemplate(&body, name, data); err != nil {
		http.Error(w, "failed to render template: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	_, _ = body.WriteTo(w)
}
