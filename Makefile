# Terraformer Makefile
#
# Wraps the exact commands run in CI (.github/workflows/{test,linter,release})
# so a green `make ci` locally means a green pipeline. Plain Go toolchain only —
# no extra deps beyond golangci-lint (installed by `make tools`).

BINARY      := terraformer
PKG         := ./...
GOFLAGS     ?=
# Pin golangci-lint to a known-good version; override on the CLI if needed.
GOLANGCI_VERSION ?= v1.64.8

# Use the repo-local module cache dir if the caller exports one (CI does not).
GO ?= go

.DEFAULT_GOAL := help

## ---------------------------------------------------------------------------
## Primary targets
## ---------------------------------------------------------------------------

.PHONY: build
build: ## Build the terraformer binary (mirrors CI `go build -v`)
	$(GO) build $(GOFLAGS) -v -o $(BINARY)

.PHONY: test
test: ## Run the full test suite (mirrors CI `go test ./...`)
	$(GO) test $(GOFLAGS) $(PKG)

.PHONY: test-race
test-race: ## Run tests with the race detector
	$(GO) test $(GOFLAGS) -race $(PKG)

.PHONY: cover
cover: ## Run tests with coverage and write coverage.out
	$(GO) test $(GOFLAGS) -coverprofile=coverage.out $(PKG)
	$(GO) tool cover -func=coverage.out | tail -1

.PHONY: lint
lint: ## Run golangci-lint (same config CI's reviewdog uses: .golangci.json)
	golangci-lint run

.PHONY: fmt
fmt: ## Format code with gofmt + goimports-style ordering
	gofmt -w .

.PHONY: fmt-check
fmt-check: ## List files that are not gofmt-clean (repo has pre-existing debt)
	@gofmt -l .

.PHONY: gcp-codegen
gcp-codegen: ## Regenerate GCP compute *_gen.go + compute.go from the Compute discovery doc (module cache)
	$(GO) run ./providers/gcp/gcp_compute_code_generator
	gofmt -w providers/gcp/*_gen.go providers/gcp/compute.go

.PHONY: vet
vet: ## Run go vet (note: `go test` also runs vet)
	$(GO) vet $(PKG)

.PHONY: tidy
tidy: ## Run go mod tidy (CI runs this — keep go.mod/go.sum clean)
	$(GO) mod tidy

.PHONY: tidy-check
tidy-check: ## Fail if go mod tidy would change go.mod/go.sum
	$(GO) mod tidy
	@git diff --exit-code go.mod go.sum || \
		(echo "go.mod/go.sum not tidy — run 'make tidy' and commit"; exit 1)

## ---------------------------------------------------------------------------
## Aggregate / CI
## ---------------------------------------------------------------------------

# `ci` mirrors .github/workflows/test.yml exactly: tidy, then build + test.
# (gofmt/vet are enforced by golangci-lint via the separate linter workflow,
# which reviewdog scopes to changed lines — a repo-wide gofmt gate would fail on
# pre-existing debt, so it is NOT part of this aggregate.)
.PHONY: ci
ci: tidy build test ## Reproduce the CI test job (workflows/test.yml)

.PHONY: check
check: lint ci ## Full local gate: linter workflow + test workflow

## ---------------------------------------------------------------------------
## Release builds (mirrors .github/workflows/release.yaml)
## ---------------------------------------------------------------------------

.PHONY: release
release: ## Cross-compile release binaries for linux/mac (amd64+arm64)
	GOOS=linux  GOARCH=amd64 $(GO) build -o $(BINARY)-all-linux-amd64
	GOOS=linux  GOARCH=arm64 $(GO) build -o $(BINARY)-all-linux-arm64
	GOOS=darwin GOARCH=amd64 $(GO) build -o $(BINARY)-all-darwin-amd64
	GOOS=darwin GOARCH=arm64 $(GO) build -o $(BINARY)-all-darwin-arm64

.PHONY: build-providers
build-providers: ## Build per-provider binaries (build/multi-build/main.go)
	$(GO) run build/multi-build/main.go

## ---------------------------------------------------------------------------
## Tooling / housekeeping
## ---------------------------------------------------------------------------

.PHONY: tools
tools: ## Install dev tools (golangci-lint) into $(go env GOPATH)/bin
	$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@$(GOLANGCI_VERSION)

.PHONY: clean
clean: ## Remove built binaries and coverage artifacts
	rm -f $(BINARY) $(BINARY)-all-* coverage.out

.PHONY: help
help: ## Show this help
	@grep -hE '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) \
		| awk 'BEGIN {FS = ":.*?## "}; {printf "  \033[36m%-16s\033[0m %s\n", $$1, $$2}'
