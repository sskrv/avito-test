package handler

import (
	"encoding/json"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"net/http"
)

type ErrorResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if data != nil {
		json.NewEncoder(w).Encode(data)
	}
}

func respondWithError(w http.ResponseWriter, statusCode int, code domain.ErrorCode, message string) {
	var errResp ErrorResponse
	errResp.Error.Code = string(code)
	errResp.Error.Message = message
	respondWithJSON(w, statusCode, errResp)
}

func handleAppError(w http.ResponseWriter, err error) {
	if appErr, ok := domain.IsAppError(err); ok {
		statusCode := http.StatusInternalServerError

		switch appErr.Code {
		case domain.ErrCodeNotFound:
			statusCode = http.StatusNotFound
		case domain.ErrCodeTeamExists, domain.ErrCodePRExists:
			statusCode = http.StatusConflict
		case domain.ErrCodePRMerged, domain.ErrCodeNotAssigned, domain.ErrCodeNoCandidate:
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusBadRequest
		}

		respondWithError(w, statusCode, appErr.Code, appErr.Message)
		return
	}

	respondWithError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
}
