package handlers

import (
	"net/http"
	"strconv"

	"exchange-rate-service/internal/models"
	"exchange-rate-service/internal/utils"
)

// CurrencyExchangeService defines the interface for currency exchange operations
// This interface allows us to keep the handler decoupled from the concrete service implementation
type CurrencyExchangeService interface {
	ConvertCurrencyAmount(fromCurrency, toCurrency string, amount float64, dateStr string) (float64, error)
	GetHistoricalExchangeRate(fromCurrency, toCurrency, dateStr string) (float64, error)
}

// ExchangeHandler handles all HTTP requests related to currency exchange
type ExchangeHandler struct {
	currencyService CurrencyExchangeService
}

// NewExchangeHandler creates a new handler instance with the provided service
func NewExchangeHandler(currencyService CurrencyExchangeService) *ExchangeHandler {
	return &ExchangeHandler{
		currencyService: currencyService,
	}
}

// Convert handles GET /convert requests
func (h *ExchangeHandler) Convert(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	// Extract required parameters
	fromCurrency := query.Get("from")
	toCurrency := query.Get("to")
	amountStr := query.Get("amount")

	// check required params
	if fromCurrency == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: from")
		return
	}
	if toCurrency == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: to")
		return
	}
	if amountStr == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: amount")
		return
	}

	// parse amount
	amount, err := strconv.ParseFloat(amountStr, 64)
	if err != nil {
		utils.ErrorResp(w, http.StatusBadRequest, "invalid amount format")
		return
	}

	// Optional date parameter
	date := query.Get("date")

	// Call our currency service to perform the conversion
	convertedAmount, err := h.currencyService.ConvertCurrencyAmount(fromCurrency, toCurrency, amount, date)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	// Build response
	response := models.ConvertResponse{
		Amount: convertedAmount,
	}

	utils.WriteJSON(w, http.StatusOK, response)
}

// latest rate endpoint
func (h *ExchangeHandler) GetLatestRate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from := q.Get("from")
	to := q.Get("to")

	// validate params
	if from == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: from")
		return
	}
	if to == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: to")
		return
	}

	// get rate by converting 1 unit
	rate, err := h.currencyService.ConvertCurrencyAmount(from, to, 1.0, "")
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	resp := models.CurrencyRate{
		From: from,
		To:   to,
		Rate: rate,
		Date: "latest",
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// historical rate handler
func (h *ExchangeHandler) GetHistoricalRate(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()

	from := q.Get("from")
	to := q.Get("to")
	dt := q.Get("date")

	// check params
	if from == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: from")
		return
	}
	if to == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: to")
		return
	}
	if dt == "" {
		utils.ErrorResp(w, http.StatusBadRequest, "missing required parameter: date")
		return
	}

	rate, err := h.currencyService.GetHistoricalExchangeRate(from, to, dt)
	if err != nil {
		h.handleServiceError(w, err)
		return
	}

	resp := models.CurrencyRate{
		From: from,
		To:   to,
		Rate: rate,
		Date: dt,
	}

	utils.WriteJSON(w, http.StatusOK, resp)
}

// map service errors to http codes
func (h *ExchangeHandler) handleServiceError(w http.ResponseWriter, err error) {
	msg := err.Error()

	switch {
	case utils.Contains(msg, "unsupported") || utils.Contains(msg, "invalid"):
		utils.ErrorResp(w, http.StatusBadRequest, msg)
	case utils.Contains(msg, "negative"):
		utils.ErrorResp(w, http.StatusBadRequest, msg)
	case utils.Contains(msg, "future") || utils.Contains(msg, "too far"):
		utils.ErrorResp(w, http.StatusBadRequest, msg)
	case utils.Contains(msg, "format"):
		utils.ErrorResp(w, http.StatusBadRequest, msg)
	case utils.Contains(msg, "api request failed") || utils.Contains(msg, "failed to fetch"):
		utils.ErrorResp(w, http.StatusServiceUnavailable, "exchange rate service temporarily unavailable")
	default:
		utils.ErrorResp(w, http.StatusInternalServerError, "internal server error")
	}
}
