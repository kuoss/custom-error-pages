GOLANGCI_LINT_VER := v1.59.1

.PHONY: lint
lint:
	go install -v github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_LINT_VER) || true
	golangci-lint run

.PHONY: test
test:
	go test --failfast ./...

.PHONY: cover
cover:
	go test -coverprofile=cover.out ./...
	go tool cover -func=cover.out

.PHONY: docker
docker:
	docker build -t custom-error-pages .
