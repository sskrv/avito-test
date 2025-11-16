package repository

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
)

type TeamRepository interface {
	CreateTeam(ctx context.Context, teamName string) error
	TeamExists(ctx context.Context, teamName string) (bool, error)
	GetTeam(ctx context.Context, teamName string) (*domain.Team, error)
}

type UserRepository interface {
	CreateOrUpdateUser(ctx context.Context, user *domain.User) error
	GetUser(ctx context.Context, userID string) (*domain.User, error)
	GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) error
	GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]domain.User, error)
}

type PullRequestRepository interface {
	CreatePR(ctx context.Context, pr *domain.PullRequest) error
	GetPR(ctx context.Context, prID string) (*domain.PullRequest, error)
	PRExists(ctx context.Context, prID string) (bool, error)
	MergePR(ctx context.Context, prID string) error
	AssignReviewer(ctx context.Context, prID, userID string) error
	UnassignReviewer(ctx context.Context, prID, userID string) error
	IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error)
	GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error)
	GetReviewersCount(ctx context.Context) (map[string]int, error)
	GetPRsCount(ctx context.Context) (int, error)
}

type Repository struct {
	Team        TeamRepository
	User        UserRepository
	PullRequest PullRequestRepository
}
