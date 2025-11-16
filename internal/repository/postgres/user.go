package postgres

import (
	"context"
	"database/sql"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
)

type UserRepo struct {
	db *sql.DB
}

func NewUserRepo(db *sql.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) CreateOrUpdateUser(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (user_id, username, team_name, is_active, updated_at)
		VALUES ($1, $2, $3, $4, NOW())
		ON CONFLICT (user_id)
		DO UPDATE SET username = $2, team_name = $3, is_active = $4, updated_at = NOW()
	`
	_, err := r.db.ExecContext(ctx, query, user.UserID, user.Username, user.TeamName, user.IsActive)
	return err
}

func (r *UserRepo) GetUser(ctx context.Context, userID string) (*domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE user_id = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, userID).Scan(
		&user.UserID,
		&user.Username,
		&user.TeamName,
		&user.IsActive,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrUserNotFound
	}
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *UserRepo) GetUsersByTeam(ctx context.Context, teamName string) ([]domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY user_id
	`

	rows, err := r.db.QueryContext(ctx, query, teamName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}

func (r *UserRepo) SetIsActive(ctx context.Context, userID string, isActive bool) error {
	query := `UPDATE users SET is_active = $1, updated_at = NOW() WHERE user_id = $2`
	result, err := r.db.ExecContext(ctx, query, isActive, userID)
	if err != nil {
		return err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return domain.ErrUserNotFound
	}

	return nil
}

func (r *UserRepo) GetActiveTeamMembers(ctx context.Context, teamName string, excludeUserID string) ([]domain.User, error) {
	query := `
		SELECT user_id, username, team_name, is_active
		FROM users
		WHERE team_name = $1 AND is_active = true AND user_id != $2
		ORDER BY user_id
	`

	rows, err := r.db.QueryContext(ctx, query, teamName, excludeUserID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var user domain.User
		if err := rows.Scan(&user.UserID, &user.Username, &user.TeamName, &user.IsActive); err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, rows.Err()
}
