.PHONY: run dev test test-race cover fmt check-fmt vet ci docker-up docker-down test-integration

TEST_DATABASE_URL ?= postgres://api:api@localhost:5432/api_test?sslmode=disable

run:
	go run ./cmd/api

dev:
	API_STORE=memory API_JWT_SECRET=dev-only-secret-change-me-32chars!! go run ./cmd/api

test:
	go test ./...

test-race:
	go test -race -count=1 ./...

test-integration:
	TEST_DATABASE_URL="$(TEST_DATABASE_URL)" go test -race -count=1 -tags integration ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

fmt:
	gofmt -w .

check-fmt:
	test -z "$$(gofmt -l .)"

vet:
	go vet ./...

ci: check-fmt vet test-race

docker-up:
	docker compose up --build

docker-down:
	docker compose down
