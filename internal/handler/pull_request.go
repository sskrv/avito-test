package handler

import (
	"encoding/json"
	"github.com/avito-test/pr-reviewer-service/internal/service"
	"net/http"
)

type PullRequestHandler struct {
	service service.PullRequestService
}

func NewPullRequestHandler(service service.PullRequestService) *PullRequestHandler {
	return &PullRequestHandler{service: service}
}

type CreatePRRequest struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
}

type MergePRRequest struct {
	PullRequestID string `json:"pull_request_id"`
}

type ReassignRequest struct {
	PullRequestID string `json:"pull_request_id"`
	OldUserID     string `json:"old_user_id"`
}

func (h *PullRequestHandler) CreatePR(w http.ResponseWriter, r *http.Request) {
	var req CreatePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, err := h.service.CreatePR(r.Context(), req.PullRequestID, req.PullRequestName, req.AuthorID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (h *PullRequestHandler) MergePR(w http.ResponseWriter, r *http.Request) {
	var req MergePRRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, err := h.service.MergePR(r.Context(), req.PullRequestID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr": pr,
	}
	respondWithJSON(w, http.StatusOK, response)
}

func (h *PullRequestHandler) Reassign(w http.ResponseWriter, r *http.Request) {
	var req ReassignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	pr, replacedBy, err := h.service.ReassignReviewer(r.Context(), req.PullRequestID, req.OldUserID)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"pr":          pr,
		"replaced_by": replacedBy,
	}
	respondWithJSON(w, http.StatusOK, response)
}
