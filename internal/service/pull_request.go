package service

import (
	"context"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/repository"
	"math/rand"
)

type pullRequestService struct {
	repo *repository.Repository
	rng  *rand.Rand
}

func NewPullRequestService(repo *repository.Repository, rng *rand.Rand) PullRequestService {
	return &pullRequestService{
		repo: repo,
		rng:  rng,
	}
}

func (s *pullRequestService) CreatePR(ctx context.Context, prID, prName, authorID string) (*domain.PullRequest, error) {
	exists, err := s.repo.PullRequest.PRExists(ctx, prID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, domain.ErrPRExists
	}

	author, err := s.repo.User.GetUser(ctx, authorID)
	if err != nil {
		return nil, domain.ErrAuthorNotFound
	}

	activeMembers, err := s.repo.User.GetActiveTeamMembers(ctx, author.TeamName, authorID)
	if err != nil {
		return nil, err
	}

	reviewers := s.selectReviewers(activeMembers, 2)

	pr := &domain.PullRequest{
		PullRequestID:     prID,
		PullRequestName:   prName,
		AuthorID:          authorID,
		Status:            domain.PRStatusOpen,
		AssignedReviewers: reviewers,
	}

	if err := s.repo.PullRequest.CreatePR(ctx, pr); err != nil {
		return nil, err
	}

	return s.repo.PullRequest.GetPR(ctx, prID)
}

func (s *pullRequestService) MergePR(ctx context.Context, prID string) (*domain.PullRequest, error) {
	pr, err := s.repo.PullRequest.GetPR(ctx, prID)
	if err != nil {
		return nil, err
	}

	if pr.Status == domain.PRStatusMerged {
		return pr, nil
	}

	if err := s.repo.PullRequest.MergePR(ctx, prID); err != nil {
		return nil, err
	}

	return s.repo.PullRequest.GetPR(ctx, prID)
}

func (s *pullRequestService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*domain.PullRequest, string, error) {
	pr, err := s.repo.PullRequest.GetPR(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	if pr.Status == domain.PRStatusMerged {
		return nil, "", domain.ErrPRMerged
	}

	isAssigned, err := s.repo.PullRequest.IsReviewerAssigned(ctx, prID, oldUserID)
	if err != nil {
		return nil, "", err
	}
	if !isAssigned {
		return nil, "", domain.ErrNotAssigned
	}

	oldReviewer, err := s.repo.User.GetUser(ctx, oldUserID)
	if err != nil {
		return nil, "", err
	}

	currentReviewers := make(map[string]bool)
	for _, reviewerID := range pr.AssignedReviewers {
		currentReviewers[reviewerID] = true
	}

	activeMembers, err := s.repo.User.GetActiveTeamMembers(ctx, oldReviewer.TeamName, "")
	if err != nil {
		return nil, "", err
	}

	var candidates []domain.User
	for _, member := range activeMembers {
		if !currentReviewers[member.UserID] && member.UserID != pr.AuthorID {
			candidates = append(candidates, member)
		}
	}

	if len(candidates) == 0 {
		return nil, "", domain.ErrNoCandidate
	}

	newReviewer := candidates[s.rng.Intn(len(candidates))]

	if err := s.repo.PullRequest.UnassignReviewer(ctx, prID, oldUserID); err != nil {
		return nil, "", err
	}

	if err := s.repo.PullRequest.AssignReviewer(ctx, prID, newReviewer.UserID); err != nil {
		return nil, "", err
	}

	updatedPR, err := s.repo.PullRequest.GetPR(ctx, prID)
	if err != nil {
		return nil, "", err
	}

	return updatedPR, newReviewer.UserID, nil
}

func (s *pullRequestService) selectReviewers(candidates []domain.User, maxCount int) []string {
	if len(candidates) == 0 {
		return []string{}
	}

	count := maxCount
	if len(candidates) < count {
		count = len(candidates)
	}

	shuffled := make([]domain.User, len(candidates))
	copy(shuffled, candidates)

	for i := len(shuffled) - 1; i > 0; i-- {
		j := s.rng.Intn(i + 1)
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	reviewers := make([]string, count)
	for i := 0; i < count; i++ {
		reviewers[i] = shuffled[i].UserID
	}

	return reviewers
}
