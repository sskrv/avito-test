package domain

import "errors"

type ErrorCode string

const (
	ErrCodeTeamExists  ErrorCode = "TEAM_EXISTS"
	ErrCodePRExists    ErrorCode = "PR_EXISTS"
	ErrCodePRMerged    ErrorCode = "PR_MERGED"
	ErrCodeNotAssigned ErrorCode = "NOT_ASSIGNED"
	ErrCodeNoCandidate ErrorCode = "NO_CANDIDATE"
	ErrCodeNotFound    ErrorCode = "NOT_FOUND"
)

type AppError struct {
	Code    ErrorCode
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

func NewAppError(code ErrorCode, message string) *AppError {
	return &AppError{
		Code:    code,
		Message: message,
	}
}

var (
	ErrTeamExists     = NewAppError(ErrCodeTeamExists, "team already exists")
	ErrPRExists       = NewAppError(ErrCodePRExists, "PR id already exists")
	ErrPRMerged       = NewAppError(ErrCodePRMerged, "cannot reassign on merged PR")
	ErrNotAssigned    = NewAppError(ErrCodeNotAssigned, "reviewer is not assigned to this PR")
	ErrNoCandidate    = NewAppError(ErrCodeNoCandidate, "no active replacement candidate in team")
	ErrTeamNotFound   = NewAppError(ErrCodeNotFound, "team not found")
	ErrUserNotFound   = NewAppError(ErrCodeNotFound, "user not found")
	ErrPRNotFound     = NewAppError(ErrCodeNotFound, "PR not found")
	ErrAuthorNotFound = NewAppError(ErrCodeNotFound, "author not found")
)

func IsAppError(err error) (*AppError, bool) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		return appErr, true
	}
	return nil, false
}
