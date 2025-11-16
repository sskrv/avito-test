package handler

import (
	"encoding/json"
	"github.com/avito-test/pr-reviewer-service/internal/service"
	"net/http"
)

type UserHandler struct {
	service service.UserService
}

func NewUserHandler(service service.UserService) *UserHandler {
	return &UserHandler{service: service}
}

type SetIsActiveRequest struct {
	UserID   string `json:"user_id"`
	IsActive bool   `json:"is_active"`
}

func (h *UserHandler) SetIsActive(w http.ResponseWriter, r *http.Request) {
	var req SetIsActiveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	user, err := h.service.SetIsActive(r.Context(), req.UserID, req.IsActive)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"user": user,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (h *UserHandler) GetReview(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("user_id")
	if userID == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "user_id is required")
		return
	}

	prs, err := h.service.GetReviewPRs(r.Context(), userID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"user_id":       userID,
		"pull_requests": prs,
	}
	respondWithJSON(w, http.StatusOK, response)
}
