.PHONY: start stop lint mocks test test-integration cover 

start:
	docker compose --project-directory . -f docker/docker-compose.yml up -d --build

stop:
	docker compose --project-directory . -f docker/docker-compose.yml down

mocks:
	mockery

lint:
	golangci-lint run

test:
	go test -race -short ./...

test-integration:
	go test -v -count=1 ./tests/...

cover:
	@go test -short -coverprofile=coverage.out ./... && \
		grep -vE '/mocks/' coverage.out > filtered.out && \
		go tool cover -func=filtered.out
	@rm -f coverage.out filtered.out