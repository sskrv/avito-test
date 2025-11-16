package service

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/repository"
)

type Statistics struct {
	TotalPRs          int            `json:"total_prs"`
	AssignmentsByUser map[string]int `json:"assignments_by_user"`
}

type statisticsService struct {
	repo *repository.Repository
}

func NewStatisticsService(repo *repository.Repository) StatisticsService {
	return &statisticsService{repo: repo}
}

func (s *statisticsService) GetStatistics(ctx context.Context) (*Statistics, error) {
	totalPRs, err := s.repo.PullRequest.GetPRsCount(ctx)
	if err != nil {
		return nil, err
	}

	assignmentsByUser, err := s.repo.PullRequest.GetReviewersCount(ctx)
	if err != nil {
		return nil, err
	}

	if assignmentsByUser == nil {
		assignmentsByUser = make(map[string]int)
	}

	return &Statistics{
		TotalPRs:          totalPRs,
		AssignmentsByUser: assignmentsByUser,
	}, nil
}
