package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/ibldzn/alma/internal/interfaces"
)

type Handler struct {
	TimeDepositService interfaces.ITimeDepositService
	SavingService      interfaces.ISavingService
	LDRService         interfaces.ILDRService
	SupermanService    interfaces.ISupermanService
}

func NewHandler(
	timeDepositService interfaces.ITimeDepositService,
	savingService interfaces.ISavingService,
	ldrService interfaces.ILDRService,
	supermanService interfaces.ISupermanService,
) *Handler {
	return &Handler{
		TimeDepositService: timeDepositService,
		SavingService:      savingService,
		LDRService:         ldrService,
		SupermanService:    supermanService,
	}
}

func (h *Handler) Router() http.Handler {
	r := chi.NewRouter()

	r.Get("/", h.Index)

	return r
}

func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	startDate := r.URL.Query().Get("start_date")
	if startDate == "" {
		startDate = "2026-06-01"
	}

	endDate := r.URL.Query().Get("end_date")
	if endDate == "" {
		endDate = "2026-06-07"
	}

	timeDeposits, err := h.TimeDepositService.GetTimeDepositSummary(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, "failed to get time deposit history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	savings, err := h.SavingService.GetSavingSummary(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, "failed to get saving history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	ldr, err := h.LDRService.GetLDRHistory(r.Context(), startDate, endDate)
	if err != nil {
		http.Error(w, "failed to get LDR history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	loanFromOtherBanks, err := h.SupermanService.GetSaldoNeracas(r.Context(), startDate, endDate, []string{"260"})
	if err != nil {
		http.Error(w, "failed to get loan from other banks history: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonResponse := struct {
		TimeDeposits       any `json:"time_deposits"`
		Savings            any `json:"savings"`
		LDR                any `json:"ldr"`
		LoanFromOtherBanks any `json:"loan_from_other_banks"`
	}{
		TimeDeposits:       timeDeposits,
		Savings:            savings,
		LDR:                ldr,
		LoanFromOtherBanks: loanFromOtherBanks,
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}
