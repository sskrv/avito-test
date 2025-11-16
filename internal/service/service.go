package service

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/repository"
	"math/rand"
	"time"
)

type TeamService interface {
	CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserService interface {
	SetIsActive(ctx context.Context, userID string, isActive bool) (*domain.User, error)
	GetReviewPRs(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
}

type PullRequestService interface {
	CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*domain.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error)
}

type StatisticsService interface {
	GetStatistics(ctx context.Context) (*Statistics, error)
}

type Service struct {
	Team        TeamService
	User        UserService
	PullRequest PullRequestService
	Statistics  StatisticsService
}

func NewService(repo *repository.Repository) *Service {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))

	return &Service{
		Team:        NewTeamService(repo),
		User:        NewUserService(repo),
		PullRequest: NewPullRequestService(repo, rng),
		Statistics:  NewStatisticsService(repo),
	}
}
