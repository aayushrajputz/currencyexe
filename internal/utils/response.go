package utils

import (
	"encoding/json"
	"log"
	"net/http"
)

// WriteJSON - helper for json responses
func WriteJSON(w http.ResponseWriter, code int, payload interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)

	if err := json.NewEncoder(w).Encode(payload); err != nil {
		log.Printf("json encode failed: %v", err)
		// fallback error
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	}
}

// send error resp
func sendErr(w http.ResponseWriter, code int, msg string) {
	errData := map[string]interface{}{
		"error":  msg,
		"status": "error",
	}
	WriteJSON(w, code, errData)
}

// success response wrapper
func SendSuccessResponse(w http.ResponseWriter, data interface{}) {
	resp := map[string]interface{}{
		"status": "success",
		"data":   data,
	}
	WriteJSON(w, http.StatusOK, resp)
}

// quick error helper - used by handlers
func ErrorResp(w http.ResponseWriter, code int, msg string) {
	sendErr(w, code, msg)
}

// Contains check - todo: maybe use strings.Contains instead?
func Contains(str, sub string) bool {
	for i := 0; i <= len(str)-len(sub); i++ {
		if str[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
