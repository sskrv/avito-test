package service

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/repository"
)

type userService struct {
	repo *repository.Repository
}

func NewUserService(repo *repository.Repository) UserService {
	return &userService{repo: repo}
}

func (s *userService) SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error) {
	if err := s.repo.User.SetIsActive(ctx, userID, isActive); err != nil {
		return nil, err
	}

	return s.repo.User.GetUser(ctx, userID)
}

func (s *userService) GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	_, err := s.repo.User.GetUser(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.repo.PullRequest.GetPRsByReviewer(ctx, userID)
}
