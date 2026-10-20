package web

import (
	"encoding/json"
	"net/http"
)

// ErrorResponse representa o envelope JSON padrão para erros da API.
type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code,omitempty"`
	Error   string `json:"error"`
	Details any    `json:"details,omitempty"`
}

// RespondJSON serializa o payload em JSON com o código de status HTTP especificado.
func RespondJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	_ = json.NewEncoder(w).Encode(payload)
}

// RespondError retorna uma resposta de erro padronizada em JSON.
func RespondError(w http.ResponseWriter, status int, code, message string, details ...any) {
	var detailVal any
	if len(details) > 0 {
		detailVal = details[0]
	}
	RespondJSON(w, status, ErrorResponse{
		Success: false,
		Code:    code,
		Error:   message,
		Details: detailVal,
	})
}

// RespondValidationError retorna um erro 400 Bad Request com código VALIDATION_ERROR e detalhes por campo.
func RespondValidationError(w http.ResponseWriter, args ...any) {
	msg := "Dados de entrada inválidos"
	var details any
	if len(args) == 1 {
		details = args[0]
	} else if len(args) >= 2 {
		if s, ok := args[0].(string); ok && s != "" {
			msg = s
		}
		details = args[1]
	}
	RespondError(w, http.StatusBadRequest, "VALIDATION_ERROR", msg, details)
}
