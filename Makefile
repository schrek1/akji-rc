# Local, manually-run tasks. Intentionally NOT wired into CI.
.PHONY: build test vet lint run

build:
	go build ./...

test:
	go test ./...

vet:
	go vet ./...

# Static analysis + linting. Run manually.
# One-time install of the tools (kept out of go.mod to keep the module dependency-free):
#   go install honnef.co/go/tools/cmd/staticcheck@latest
#   go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
lint: vet
	staticcheck ./...
	golangci-lint run

run:
	go run ./cmd/akji-capture
