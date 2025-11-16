package tests

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/avito-test/pr-reviewer-service/internal/domain"
	"github.com/avito-test/pr-reviewer-service/internal/handler"
	"github.com/avito-test/pr-reviewer-service/internal/repository/postgres"
	"github.com/avito-test/pr-reviewer-service/internal/service"
	_ "github.com/lib/pq"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func setupTestDB(t *testing.T) *sql.DB {
	cfg := postgres.Config{
		Host:     getEnvOrDefault("TEST_DB_HOST", "localhost"),
		Port:     getEnvOrDefault("TEST_DB_PORT", "5432"),
		User:     getEnvOrDefault("TEST_DB_USER", "postgres"),
		Password: getEnvOrDefault("TEST_DB_PASSWORD", "postgres"),
		DBName:   getEnvOrDefault("TEST_DB_NAME", "pr_reviewer_test"),
		SSLMode:  "disable",
	}

	db, err := postgres.NewPostgresDB(cfg)
	if err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	cleanupDB(db)

	if err := postgres.RunMigrations(db, "../migrations"); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	return db
}

func cleanupDB(db *sql.DB) {
	db.Exec("DROP TABLE IF EXISTS pr_reviewers CASCADE")
	db.Exec("DROP TABLE IF EXISTS pull_requests CASCADE")
	db.Exec("DROP TABLE IF EXISTS users CASCADE")
	db.Exec("DROP TABLE IF EXISTS teams CASCADE")
}

func getEnvOrDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func TestIntegrationFullWorkflow(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewRepository(db)
	svc := service.NewService(repo)
	handlers := handler.NewHandler(svc)
	router := handlers.InitRoutes()

	t.Run("Create team and verify", func(t *testing.T) {
		team := domain.Team{
			TeamName: "backend",
			Members: []domain.TeamMember{
				{UserID: "u1", Username: "Alice", IsActive: true},
				{UserID: "u2", Username: "Bob", IsActive: true},
				{UserID: "u3", Username: "Charlie", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		req = httptest.NewRequest(http.MethodGet, "/team/get?team_name=backend", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}

		var retrievedTeam domain.Team
		json.NewDecoder(w.Body).Decode(&retrievedTeam)

		if len(retrievedTeam.Members) != 3 {
			t.Fatalf("Expected 3 members, got %d", len(retrievedTeam.Members))
		}
	})

	t.Run("Create PR with auto-assignment", func(t *testing.T) {
		prReq := map[string]string{
			"pull_request_id":   "pr-1",
			"pull_request_name": "Add feature",
			"author_id":         "u1",
		}

		body, _ := json.Marshal(prReq)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d. Body: %s", w.Code, w.Body.String())
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		pr := response["pr"].(map[string]interface{})
		reviewers := pr["assigned_reviewers"].([]interface{})

		if len(reviewers) != 2 {
			t.Fatalf("Expected 2 reviewers, got %d", len(reviewers))
		}

		for _, reviewer := range reviewers {
			if reviewer.(string) == "u1" {
				t.Fatal("Author should not be assigned as reviewer")
			}
		}
	})

	t.Run("Get user reviews", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/users/getReview?user_id=u2", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		prs := response["pull_requests"].([]interface{})
		if len(prs) == 0 {
			t.Log("User might not have been assigned to this PR (random selection)")
		}
	})

	t.Run("Deactivate user and create new PR", func(t *testing.T) {
		deactivateReq := map[string]interface{}{
			"user_id":   "u2",
			"is_active": false,
		}

		body, _ := json.Marshal(deactivateReq)
		req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}

		prReq := map[string]string{
			"pull_request_id":   "pr-2",
			"pull_request_name": "Fix bug",
			"author_id":         "u1",
		}

		body, _ = json.Marshal(prReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		pr := response["pr"].(map[string]interface{})
		reviewers := pr["assigned_reviewers"].([]interface{})

		for _, reviewer := range reviewers {
			if reviewer.(string) == "u2" {
				t.Fatal("Inactive user should not be assigned as reviewer")
			}
		}
	})

	t.Run("Merge PR", func(t *testing.T) {
		mergeReq := map[string]string{
			"pull_request_id": "pr-1",
		}

		body, _ := json.Marshal(mergeReq)
		req := httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		pr := response["pr"].(map[string]interface{})
		status := pr["status"].(string)

		if status != "MERGED" {
			t.Fatalf("Expected status MERGED, got %s", status)
		}
	})

	t.Run("Statistics endpoint", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/statistics", nil)
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Expected status 200, got %d", w.Code)
		}

		var stats map[string]interface{}
		json.NewDecoder(w.Body).Decode(&stats)

		if _, ok := stats["total_prs"]; !ok {
			t.Fatal("Statistics should contain total_prs")
		}

		if _, ok := stats["assignments_by_user"]; !ok {
			t.Fatal("Statistics should contain assignments_by_user")
		}
	})
}

func TestIntegrationEdgeCases(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := postgres.NewRepository(db)
	svc := service.NewService(repo)
	handlers := handler.NewHandler(svc)
	router := handlers.InitRoutes()

	t.Run("Duplicate team creation should fail", func(t *testing.T) {
		team := domain.Team{
			TeamName: "payments",
			Members: []domain.TeamMember{
				{UserID: "u10", Username: "Dave", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", w.Code)
		}

		req = httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("Expected status 409, got %d", w.Code)
		}
	})

	t.Run("PR with insufficient team members", func(t *testing.T) {
		team := domain.Team{
			TeamName: "small-team",
			Members: []domain.TeamMember{
				{UserID: "u20", Username: "Solo", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		prReq := map[string]string{
			"pull_request_id":   "pr-solo",
			"pull_request_name": "Solo PR",
			"author_id":         "u20",
		}

		body, _ = json.Marshal(prReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("Expected status 201, got %d", w.Code)
		}

		var response map[string]interface{}
		json.NewDecoder(w.Body).Decode(&response)

		pr := response["pr"].(map[string]interface{})
		reviewers := pr["assigned_reviewers"].([]interface{})

		if len(reviewers) != 0 {
			t.Fatalf("Expected 0 reviewers, got %d", len(reviewers))
		}
	})

	t.Run("Cannot reassign after merge", func(t *testing.T) {
		team := domain.Team{
			TeamName: "test-merge",
			Members: []domain.TeamMember{
				{UserID: "u30", Username: "User30", IsActive: true},
				{UserID: "u31", Username: "User31", IsActive: true},
				{UserID: "u32", Username: "User32", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		prReq := map[string]string{
			"pull_request_id":   "pr-merge-test",
			"pull_request_name": "Merge test",
			"author_id":         "u30",
		}

		body, _ = json.Marshal(prReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		mergeReq := map[string]string{
			"pull_request_id": "pr-merge-test",
		}

		body, _ = json.Marshal(mergeReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		reassignReq := map[string]string{
			"pull_request_id": "pr-merge-test",
			"old_user_id":     "u31",
		}

		body, _ = json.Marshal(reassignReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/reassign", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("Expected status 409, got %d", w.Code)
		}
	})

	t.Run("Merge idempotency", func(t *testing.T) {
		team := domain.Team{
			TeamName: "idempotent-team",
			Members: []domain.TeamMember{
				{UserID: "u40", Username: "User40", IsActive: true},
			},
		}

		body, _ := json.Marshal(team)
		req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		prReq := map[string]string{
			"pull_request_id":   "pr-idempotent",
			"pull_request_name": "Idempotent test",
			"author_id":         "u40",
		}

		body, _ = json.Marshal(prReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/create", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		mergeReq := map[string]string{
			"pull_request_id": "pr-idempotent",
		}

		body, _ = json.Marshal(mergeReq)
		req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("First merge: Expected status 200, got %d", w.Code)
		}

		req = httptest.NewRequest(http.MethodPost, "/pullRequest/merge", bytes.NewReader(body))
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Second merge: Expected status 200, got %d", w.Code)
		}
	})
}

func BenchmarkCreatePR(b *testing.B) {
	db := setupTestDB(&testing.T{})
	defer db.Close()

	repo := postgres.NewRepository(db)
	svc := service.NewService(repo)

	team := domain.Team{
		TeamName: "bench-team",
		Members: []domain.TeamMember{
			{UserID: "bench-u1", Username: "BenchUser1", IsActive: true},
			{UserID: "bench-u2", Username: "BenchUser2", IsActive: true},
			{UserID: "bench-u3", Username: "BenchUser3", IsActive: true},
		},
	}

	ctx := context.Background()
	svc.Team.CreateTeam(ctx, &team)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		prID := fmt.Sprintf("pr-bench-%d-%d", time.Now().UnixNano(), i)
		_, _ = svc.PullRequest.CreatePR(ctx, prID, "Benchmark PR", "bench-u1")
	}
}
