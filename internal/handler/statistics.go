package handler

import (
	"github.com/avito-test/pr-reviewer-service/internal/service"
	"net/http"
)

type StatisticsHandler struct {
	service service.StatisticsService
}

func NewStatisticsHandler(service service.StatisticsService) *StatisticsHandler {
	return &StatisticsHandler{service: service}
}

func (h *StatisticsHandler) GetStatistics(w http.ResponseWriter, r *http.Request) {
	stats, err := h.service.GetStatistics(r.Context())
	if err != nil {
		handleAppError(w, err)
		return
	}

	respondWithJSON(w, http.StatusOK, stats)
}
