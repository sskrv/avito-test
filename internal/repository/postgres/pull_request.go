package postgres

import (
	"context"
	"database/sql"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"time"
)

type PullRequestRepo struct {
	db *sql.DB
}

func NewPullRequestRepo(db *sql.DB) *PullRequestRepo {
	return &PullRequestRepo{db: db}
}

func (r *PullRequestRepo) CreatePR(ctx context.Context, pr *domain.PullRequest) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	query := `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status)
		VALUES ($1, $2, $3, $4)
	`
	_, err = tx.ExecContext(ctx, query, pr.PullRequestID, pr.PullRequestName, pr.AuthorID, pr.Status)
	if err != nil {
		return err
	}

	for _, reviewerID := range pr.AssignedReviewers {
		reviewerQuery := `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2)`
		_, err = tx.ExecContext(ctx, reviewerQuery, pr.PullRequestID, reviewerID)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

func (r *PullRequestRepo) GetPR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	query := `
		SELECT pull_request_id, pull_request_name, author_id, status, created_at, merged_at
		FROM pull_requests
		WHERE pull_request_id = $1
	`

	var pr domain.PullRequest
	var createdAt time.Time
	var mergedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, prID).Scan(
		&pr.PullRequestID,
		&pr.PullRequestName,
		&pr.AuthorID,
		&pr.Status,
		&createdAt,
		&mergedAt,
	)

	if err == sql.ErrNoRows {
		return nil, domain.ErrPRNotFound
	}
	if err != nil {
		return nil, err
	}

	pr.CreatedAt = &createdAt
	if mergedAt.Valid {
		pr.MergedAt = &mergedAt.Time
	}

	reviewersQuery := `SELECT user_id FROM pr_reviewers WHERE pull_request_id = $1 ORDER BY user_id`
	rows, err := r.db.QueryContext(ctx, reviewersQuery, prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	pr.AssignedReviewers = []string{}
	for rows.Next() {
		var reviewerID string
		if err := rows.Scan(&reviewerID); err != nil {
			return nil, err
		}
		pr.AssignedReviewers = append(pr.AssignedReviewers, reviewerID)
	}

	return &pr, rows.Err()
}

func (r *PullRequestRepo) PRExists(ctx context.Context, prID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID).Scan(&exists)
	return exists, err
}

func (r *PullRequestRepo) MergePR(ctx context.Context, prID string) error {
	query := `UPDATE pull_requests SET status = $1, merged_at = NOW() WHERE pull_request_id = $2 AND status != $1`
	_, err := r.db.ExecContext(ctx, query, domain.PRStatusMerged, prID)
	return err
}

func (r *PullRequestRepo) AssignReviewer(ctx context.Context, prID, userID string) error {
	query := `INSERT INTO pr_reviewers (pull_request_id, user_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.ExecContext(ctx, query, prID, userID)
	return err
}

func (r *PullRequestRepo) UnassignReviewer(ctx context.Context, prID, userID string) error {
	query := `DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2`
	_, err := r.db.ExecContext(ctx, query, prID, userID)
	return err
}

func (r *PullRequestRepo) IsReviewerAssigned(ctx context.Context, prID, userID string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pull_request_id = $1 AND user_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, prID, userID).Scan(&exists)
	return exists, err
}

func (r *PullRequestRepo) GetPRsByReviewer(ctx context.Context, userID string) ([]domain.PullRequestShort, error) {
	query := `
		SELECT pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		INNER JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.user_id = $1
		ORDER BY pr.created_at DESC
	`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var prs []domain.PullRequestShort
	for rows.Next() {
		var pr domain.PullRequestShort
		if err := rows.Scan(&pr.PullRequestID, &pr.PullRequestName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, err
		}
		prs = append(prs, pr)
	}

	if prs == nil {
		prs = []domain.PullRequestShort{}
	}

	return prs, rows.Err()
}

func (r *PullRequestRepo) GetReviewersCount(ctx context.Context) (map[string]int, error) {
	query := `
		SELECT user_id, COUNT(*) as count
		FROM pr_reviewers
		GROUP BY user_id
	`

	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]int)
	for rows.Next() {
		var userID string
		var count int
		if err := rows.Scan(&userID, &count); err != nil {
			return nil, err
		}
		result[userID] = count
	}

	return result, rows.Err()
}

func (r *PullRequestRepo) GetPRsCount(ctx context.Context) (int, error) {
	query := `SELECT COUNT(*) FROM pull_requests`
	var count int
	err := r.db.QueryRowContext(ctx, query).Scan(&count)
	return count, err
}
