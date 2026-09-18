# Host and cross builds, run, clean.

##@ Build
.PHONY: build build-all run clean
build: ## Build a static binary for this host into bin/
	CGO_ENABLED=0 $(GO) build $(BUILD_FLAGS) -o $(BIN_DIR)/ ./cmd/$(APP)

build-all: ## Cross-build static binaries for PLATFORMS into bin/<app>-<os>-<arch>
	@set -e; for p in $(PLATFORMS); do \
	  os=$${p%/*}; arch=$${p#*/}; out=$(BIN_DIR)/$(APP)-$$os-$$arch; \
	  if [ "$$os" = windows ]; then out=$$out.exe; fi; \
	  echo "GOOS=$$os GOARCH=$$arch $(GO) build -o $$out"; \
	  CGO_ENABLED=0 GOOS=$$os GOARCH=$$arch $(GO) build $(BUILD_FLAGS) -o $$out ./cmd/$(APP); \
	done

ARGS ?= version
run: ## Run the command from source with ARGS (default: version)
	$(GO) run ./cmd/$(APP) $(ARGS)

clean: ## Remove build, release and coverage outputs
	rm -rf $(BIN_DIR) $(DIST_DIR) coverage.out
