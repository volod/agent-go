# Reproducible release archives from the cross-built binaries.

##@ Distribution
.PHONY: dist
dist: build-all ## Package archives and SHA256SUMS into dist/; DIST_VERSION=vX.Y.Z requires that version
	rm -rf $(DIST_DIR)
	$(GO) run ./tools/dist -app $(APP) -bin $(BIN_DIR) -out $(DIST_DIR) -files "$(DIST_FILES)" \
	  $(if $(DIST_VERSION),-version $(DIST_VERSION)) $(PLATFORMS)
