.PHONY: build run test docker-up docker-down docker-build clean lint

build:
	go build -o server ./cmd/server

run:
	go run ./cmd/server

test:
	go test -v -race ./...

docker-build:
	docker-compose build

docker-up:
	docker-compose up -d

docker-down:
	docker-compose down

docker-logs:
	docker-compose logs -f app

clean:
	rm -f server
	docker-compose down -v

lint:
	golangci-lint run

integration-test:
	go test -v -race -tags=integration ./tests/...
