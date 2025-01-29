.PHONY: build
build:
	CGO_ENABLED=1 go build shift-shift.go

.PHONY: lint
lint:
	golangci-lint run ./...
