# Formatting, static analysis, tests, the CI gate and planning checks.
GO_DIRS = $$($(GO) list -f '{{.Dir}}' ./...)

##@ Quality
.PHONY: fmt fmt-check vet lint tidy-check test test-race coverage ci
fmt: ## Format Go sources with gofmt -s
	gofmt -s -w $(GO_DIRS)

fmt-check: ## Fail when a Go source is not gofmt -s formatted
	@out=$$(gofmt -s -l $(GO_DIRS)); \
	if [ -n "$$out" ]; then echo "unformatted (run make fmt):"; echo "$$out"; exit 1; fi

vet: ## Run go vet
	$(GO) vet ./...

lint: ## Run staticcheck, pinned by the tool directive in go.mod
	$(GO) tool staticcheck ./...

tidy-check: ## Fail when go.mod or go.sum is not tidy
	$(GO) mod tidy -diff

test: ## Run unit and integration tests
	$(GO) test ./...

test-race: ## Run tests with the race detector (needs cgo and a C compiler)
	CGO_ENABLED=1 $(GO) test -race ./...

coverage: ## Write a diagnostic coverage report (never a gate)
	$(GO) test -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out | tail -n 1

ci: fmt-check vet lint tidy-check test build-all lint-spec-plan lint-doc-links ## Required gate before accepting a task; CI runs it

##@ Planning
.PHONY: plan-status lint-spec-plan lint-doc-links
plan-status: ## Report open tasks and the next eligible task
	@$(GO) run ./tools/plancheck status

lint-spec-plan: ## Check that the capability registry, plan and records agree
	$(GO) run ./tools/plancheck lint

lint-doc-links: ## Check relative Markdown links and anchors
	$(GO) run ./tools/plancheck links
