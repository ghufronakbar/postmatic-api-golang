# =========================================
# Makefile - postmatic-api
# =========================================

APP_NAME      := postmatic-api
CMD_DIR       := ./cmd/api
BIN_DIR       := ./bin
BIN_PATH      := $(BIN_DIR)/$(APP_NAME)

GO           ?= go
GOFLAGS      ?=
GOLANGCI_LINT ?= golangci-lint
GOOSE        ?= goose

# kalau kamu pakai file env:
ENV_FILE     ?= .env

# default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  make tidy        - go mod tidy"
	@echo "  make fmt         - gofmt (write)"
	@echo "  make fmt-check   - gofmt (check only)"
	@echo "  make vet         - go vet"
	@echo "  make test        - unit tests"
	@echo "  make test-race   - unit tests with race"
	@echo "  make build       - build binary"
	@echo "  make run         - run cmd/api"
	@echo "  make clean       - remove bin/"
	@echo "  make lint        - golangci-lint run"
	@echo "  make lint-fix    - golangci-lint with --fix (limited)"
	@echo "  make ci          - fmt-check + test + lint"
	@echo ""
	@echo "Migrations (optional):"
	@echo "  make migrate-up       - goose up"
	@echo "  make migrate-down     - goose down"
	@echo "  make migrate-status   - goose status"
	@echo ""
	@echo "Env:"
	@echo "  ENV_FILE=$(ENV_FILE)"

# -------------------------
# Go modules
# -------------------------
.PHONY: tidy
tidy:
	$(GO) mod tidy

# -------------------------
# Formatting
# -------------------------
.PHONY: fmt
fmt:
	@echo "Running gofmt..."
	@gofmt -w .

.PHONY: fmt-check
fmt-check:
	@echo "Checking gofmt..."
	@unformatted=$$(gofmt -l .); \
	if [ -n "$$unformatted" ]; then \
		echo "These files are not gofmt-ed:"; \
		echo "$$unformatted"; \
		exit 1; \
	fi

# -------------------------
# Vet & tests
# -------------------------
.PHONY: vet
vet:
	$(GO) vet ./...

.PHONY: test
test:
	$(GO) test ./... -count=1

.PHONY: test-race
test-race:
	$(GO) test ./... -count=1 -race

# -------------------------
# Build & run
# -------------------------
.PHONY: build
build:
	@mkdir -p $(BIN_DIR)
	$(GO) build $(GOFLAGS) -o $(BIN_PATH) $(CMD_DIR)

.PHONY: run
run:
	@# load env from ENV_FILE if exists (mac/linux)
	@if [ -f "$(ENV_FILE)" ]; then \
		set -a; . "$(ENV_FILE)"; set +a; \
	fi; \
	$(GO) run $(CMD_DIR)

.PHONY: clean
clean:
	rm -rf $(BIN_DIR)

# -------------------------
# Linting (golangci-lint)
# -------------------------
.PHONY: lint
lint:
	@$(GOLANGCI_LINT) version >/dev/null 2>&1 || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT) run ./...

.PHONY: lint-fix
lint-fix:
	@$(GOLANGCI_LINT) version >/dev/null 2>&1 || (echo "golangci-lint not found. Install: https://golangci-lint.run/usage/install/" && exit 1)
	$(GOLANGCI_LINT) run --fix ./...

# -------------------------
# CI aggregate
# -------------------------
.PHONY: ci
ci: fmt-check test lint

# -------------------------
# Migrations (optional - Goose)
# Requires:
#  - goose installed
#  - DATABASE_URL set in env or ENV_FILE
#  - migrations directory at ./migrations
# -------------------------
MIGRATIONS_DIR := ./migrations

.PHONY: migrate-up
migrate-up:
	@$(GOOSE) -version >/dev/null 2>&1 || (echo "goose not found. Install: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	@if [ -f "$(ENV_FILE)" ]; then set -a; . "$(ENV_FILE)"; set +a; fi; \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" up

.PHONY: migrate-down
migrate-down:
	@$(GOOSE) -version >/dev/null 2>&1 || (echo "goose not found. Install: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	@if [ -f "$(ENV_FILE)" ]; then set -a; . "$(ENV_FILE)"; set +a; fi; \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" down

.PHONY: migrate-status
migrate-status:
	@$(GOOSE) -version >/dev/null 2>&1 || (echo "goose not found. Install: go install github.com/pressly/goose/v3/cmd/goose@latest" && exit 1)
	@if [ -f "$(ENV_FILE)" ]; then set -a; . "$(ENV_FILE)"; set +a; fi; \
	$(GOOSE) -dir $(MIGRATIONS_DIR) postgres "$$DATABASE_URL" status
