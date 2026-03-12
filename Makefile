.PHONY: test cover bench lint golden-update ci build

# Run all tests
test:
	go test ./... -v -count=1

# Run tests with coverage report
cover:
	go test ./... -coverprofile=coverage.out -count=1
	go tool cover -func=coverage.out
	@echo ""
	@echo "To view HTML report: go tool cover -html=coverage.out"

# Run benchmarks
bench:
	go test ./pkg/enhancer/ -bench=. -benchmem -count=3

# Run go vet
lint:
	go vet ./...

# Update golden files
golden-update:
	GOLDEN_UPDATE=1 go test ./pkg/enhancer/ -run TestGolden -v

# Full CI pipeline
ci: lint test cover bench
	@echo "CI passed"

# Build binary
build:
	go build -o prompt-improver .
