#!/bin/bash

BASE_URL="http://localhost:8080"

echo "=== Health Check ==="
curl -X GET "$BASE_URL/health"
echo -e "\n"

echo "=== Create Team 'backend' ==="
curl -X POST "$BASE_URL/team/add" \
  -H "Content-Type: application/json" \
  -d '{
    "team_name": "backend",
    "members": [
      {"user_id": "u1", "username": "Alice", "is_active": true},
      {"user_id": "u2", "username": "Bob", "is_active": true},
      {"user_id": "u3", "username": "Charlie", "is_active": true}
    ]
  }'
echo -e "\n"

echo "=== Get Team 'backend' ==="
curl -X GET "$BASE_URL/team/get?team_name=backend"
echo -e "\n"

echo "=== Create PR ==="
curl -X POST "$BASE_URL/pullRequest/create" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "pull_request_name": "Add search feature",
    "author_id": "u1"
  }'
echo -e "\n"

echo "=== Get User Reviews ==="
curl -X GET "$BASE_URL/users/getReview?user_id=u2"
echo -e "\n"

echo "=== Deactivate User ==="
curl -X POST "$BASE_URL/users/setIsActive" \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "u2",
    "is_active": false
  }'
echo -e "\n"

echo "=== Create Another PR ==="
curl -X POST "$BASE_URL/pullRequest/create" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1002",
    "pull_request_name": "Fix bug",
    "author_id": "u1"
  }'
echo -e "\n"

echo "=== Reassign Reviewer ==="
curl -X POST "$BASE_URL/pullRequest/reassign" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u3"
  }'
echo -e "\n"

echo "=== Merge PR ==="
curl -X POST "$BASE_URL/pullRequest/merge" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
echo -e "\n"

echo "=== Merge PR Again (Idempotency Test) ==="
curl -X POST "$BASE_URL/pullRequest/merge" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001"
  }'
echo -e "\n"

echo "=== Try to Reassign After Merge (Should Fail) ==="
curl -X POST "$BASE_URL/pullRequest/reassign" \
  -H "Content-Type: application/json" \
  -d '{
    "pull_request_id": "pr-1001",
    "old_user_id": "u2"
  }'
echo -e "\n"

echo "=== Get Statistics ==="
curl -X GET "$BASE_URL/statistics"
echo -e "\n"
