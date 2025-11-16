package handler

import (
	"github.com/avito-test/pr-reviewer-service/internal/service"
	"net/http"
)

type Handler struct {
	Team        *TeamHandler
	User        *UserHandler
	PullRequest *PullRequestHandler
	Statistics  *StatisticsHandler
}

func NewHandler(service *service.Service) *Handler {
	return &Handler{
		Team:        NewTeamHandler(service.Team),
		User:        NewUserHandler(service.User),
		PullRequest: NewPullRequestHandler(service.PullRequest),
		Statistics:  NewStatisticsHandler(service.Statistics),
	}
}

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("/team/add", h.Team.CreateTeam)
	mux.HandleFunc("/team/get", h.Team.GetTeam)

	mux.HandleFunc("/users/setIsActive", h.User.SetIsActive)
	mux.HandleFunc("/users/getReview", h.User.GetReview)

	mux.HandleFunc("/pullRequest/create", h.PullRequest.CreatePR)
	mux.HandleFunc("/pullRequest/merge", h.PullRequest.MergePR)
	mux.HandleFunc("/pullRequest/reassign", h.PullRequest.Reassign)

	mux.HandleFunc("/statistics", h.Statistics.GetStatistics)

	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		respondWithJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	return mux
}
