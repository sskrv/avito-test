package service

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/repository"
)

type teamService struct {
	repo *repository.Repository
}

func NewTeamService(repo *repository.Repository) TeamService {
	return &teamService{repo: repo}
}

func (s *teamService) CreateTeam(ctx context.Context, team *domain.Team) (*domain.Team, error) {
	exists, err := s.repo.Team.TeamExists(ctx, team.TeamName)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrTeamExists
	}

	if err := s.repo.Team.CreateTeam(ctx, team.TeamName); err != nil {
		return nil, err
	}

	for _, member := range team.Members {
		user := &domain.User{
			UserID:   member.UserID,
			Username: member.Username,
			TeamName: team.TeamName,
			IsActive: member.IsActive,
		}
		if err := s.repo.User.CreateOrUpdateUser(ctx, user); err != nil {
			return nil, err
		}
	}

	return s.repo.Team.GetTeam(ctx, team.TeamName)
}

func (s *teamService) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	return s.repo.Team.GetTeam(ctx, teamName)
}
