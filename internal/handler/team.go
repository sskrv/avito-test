package handler

import (
	"encoding/json"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/service"
	"net/http"
)

type TeamHandler struct {
	service service.TeamService
}

func NewTeamHandler(service service.TeamService) *TeamHandler {
	return &TeamHandler{service: service}
}

func (h *TeamHandler) CreateTeam(w http.ResponseWriter, r *http.Request) {
	var team domain.Team
	if err := json.NewDecoder(r.Body).Decode(&team); err != nil {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "invalid request body")
		return
	}

	if team.Members == nil {
		team.Members = []domain.TeamMember{}
	}

	createdTeam, err := h.service.CreateTeam(r.Context(), &team)
	if err != nil {
		handleAppError(w, err)
		return
	}

	response := map[string]interface{}{
		"team": createdTeam,
	}
	respondWithJSON(w, http.StatusCreated, response)
}

func (h *TeamHandler) GetTeam(w http.ResponseWriter, r *http.Request) {
	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		respondWithError(w, http.StatusBadRequest, "INVALID_REQUEST", "team_name is required")
		return
	}

	team, err := h.service.GetTeam(r.Context(), teamName)
	if err != nil {
		handleAppError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, team)
}
