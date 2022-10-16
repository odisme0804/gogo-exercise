.PHONY: dev
dev:
	go run ./cmd/app/main.go

.PHONY: build
build:
	$(shell ls cmd | xargs -I {} go build -o bin/{} cmd/{}/main.go)

.PHONY: docker-build
docker-build:
	docker build -t gogo-exercise . --no-cache

.PHONY: run
run: build
	./bin/app

.PHONY: test
test:
	go test -race -cover -coverprofile cover.out ./...
	go tool cover -func=cover.out | tail -n 1 | awk '{print $3}'
