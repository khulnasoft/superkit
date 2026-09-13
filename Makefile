TEMP_DIR := ./.tmp

# Tool versions #################################
GOLICENSES_VERSION := v5.0.1

# Formatting variables #################################
BOLD := $(shell tput -T linux bold)
PURPLE := $(shell tput -T linux setaf 5)
GREEN := $(shell tput -T linux setaf 2)
CYAN := $(shell tput -T linux setaf 6)
RED := $(shell tput -T linux setaf 1)
RESET := $(shell tput -T linux sgr0)
TITLE := $(BOLD)$(PURPLE)
SUCCESS := $(BOLD)$(GREEN)

define title
	@printf "$(TITLE)%s$(RESET)\n" "$(1)"
endef

# Test variables #################################
COVERAGE_THRESHOLD := 28  # the quality gate lower threshold for unit test total % coverage (by function statements)


## Bootstrapping targets #################################

.PHONY: bootstrap
bootstrap: $(TEMP_DIR) bootstrap-go bootstrap-tools ## Download and install all tooling dependencies (+ prep tooling in the ./tmp dir)
	$(call title,Bootstrapping dependencies)

.PHONY: bootstrap-go
bootstrap-go:
	go mod download

$(TEMP_DIR):
	mkdir -p $@

.PHONY: bootstrap-tools
bootstrap-tools: $(TEMP_DIR)
	GOBIN=$(realpath $(TEMP_DIR)) go install golang.org/x/perf/cmd/benchstat@latest
	curl -sSfL https://raw.githubusercontent.com/khulnasoft/go-licenses/master/golicenses.sh | sh -s -- -b $(TEMP_DIR)/ $(GOLICENSES_VERSION)

.PHONY: static-analysis
static-analysis: check-go-mod-tidy check-licenses  ## Run all static analysis checks

.PHONY: check-licenses
check-licenses:  ## Ensure transitive dependencies are compliant with the current license policy
	$(call title,Checking for license compliance)
	$(TEMP_DIR)/golicenses check ./...


## Testing targets #################################

.PHONY: unit
unit: $(TEMP_DIR)  ## Run unit tests (with coverage)
	$(call title,Running unit tests)
	go test -coverprofile $(TEMP_DIR)/unit-coverage-details.txt ./...
	@.github/scripts/coverage.py $(COVERAGE_THRESHOLD) $(TEMP_DIR)/unit-coverage-details.txt

.PHONY: check-go-mod-tidy
check-go-mod-tidy:
	@ .github/scripts/go-mod-tidy-check.sh && echo "go.mod and go.sum are tidy!"


## Application targets #################################

.PHONY: test
test:  ## Run root module tests
	@go test -v ./...

.PHONY: app-test
app-test:  ## Run bootstrap application tests
	@cd bootstrap && go test -v ./...

.PHONY: app-generate
app-generate:  ## Generate Templ files for bootstrap app
	@cd bootstrap && go run github.com/a-h/templ/cmd/templ@v0.3.865 generate

.PHONY: app-build
app-build: app-generate  ## Build bootstrap application for production
	@cd bootstrap && make build

.PHONY: app-dev
app-dev:  ## Run bootstrap app in development mode (from bootstrap directory)
	@cd bootstrap && make dev


## Halp! #################################

.PHONY: help
help:  ## Display this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "$(BOLD)$(CYAN)%-25s$(RESET)%s\n", $$1, $$2}'
