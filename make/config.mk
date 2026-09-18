# Shared variables. Personalizing the template changes APP.
GO          ?= go
APP         := agent-go
BIN_DIR     := bin
# Cross-build targets for `make build-all`; gates run on the host only.
PLATFORMS   ?= linux/amd64 linux/arm64 darwin/arm64 windows/amd64
# Reproducible, stripped binaries; builds also set CGO_ENABLED=0 for static linking.
# The version comes from Go's VCS stamping, not from ldflags.
BUILD_FLAGS := -trimpath -ldflags "-s -w"
# Release archives: one per platform plus SHA256SUMS, each holding the binary and DIST_FILES.
DIST_DIR    := dist
DIST_FILES  := README.md LICENSE
