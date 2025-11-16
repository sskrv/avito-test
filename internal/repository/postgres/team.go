package postgres

import (
	"context"
	"database/sql"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
)

type TeamRepo struct {
	db *sql.DB
}

func NewTeamRepo(db *sql.DB) *TeamRepo {
	return &TeamRepo{db: db}
}

func (r *TeamRepo) CreateTeam(ctx context.Context, teamName string) error {
	query := `INSERT INTO teams (team_name) VALUES ($1)`
	_, err := r.db.ExecContext(ctx, query, teamName)
	return err
}

func (r *TeamRepo) TeamExists(ctx context.Context, teamName string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM teams WHERE team_name = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, teamName).Scan(&exists)
	return exists, err
}

func (r *TeamRepo) GetTeam(ctx context.Context, teamName string) (*domain.Team, error) {
	exists, err := r.TeamExists(ctx, teamName)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, domain.ErrTeamNotFound
	}

	query := `
		SELECT u.user_id, u.username, u.is_active
		FROM users u
		WHERE u.team_name = $1
		ORDER BY u.user_id
	`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var members []domain.TeamMember
	for rows.Next() {
		var member domain.TeamMember
		if err := rows.Scan(&member.UserID, &member.Username, &member.IsActive); err != nil {
			return nil, err
		}
		members = append(members, member)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	if members == nil {
		members = []domain.TeamMember{}
	}

	return &domain.Team{
		TeamName: teamName,
		Members:  members,
	}, nil
}
