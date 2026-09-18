package buildinfo

import (
	"runtime/debug"
	"strings"
	"testing"
)

func TestFromBuildInfo(t *testing.T) {
	build := func(path, version string) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Path: path, Version: version}}
	}
	cases := []struct {
		name        string
		bi          *debug.BuildInfo
		wantModule  string
		wantVersion string
	}{
		{"no build info", nil, "", DevelVersion},
		{"go run or go test", build("example.com/app", "(devel)"), "example.com/app", DevelVersion},
		{"empty version", build("example.com/app", ""), "example.com/app", DevelVersion},
		{"tagged build", build("example.com/app", "v1.2.3"), "example.com/app", "v1.2.3"},
		{"untagged build", build("example.com/app", "v0.0.0-20260918045403-c68e0f2e8a32+dirty"),
			"example.com/app", "v0.0.0-20260918045403-c68e0f2e8a32+dirty"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := fromBuildInfo(tc.bi)
			if got.Name != Name || got.Module != tc.wantModule || got.Version != tc.wantVersion {
				t.Fatalf("fromBuildInfo() = %+v, want name %q, module %q, version %q",
					got, Name, tc.wantModule, tc.wantVersion)
			}
			if got.GoVersion == "" || !strings.Contains(got.Platform, "/") {
				t.Fatalf("fromBuildInfo() = %+v, want Go version and GOOS/GOARCH", got)
			}
		})
	}
}

func TestInfoString(t *testing.T) {
	info := Info{Name: "app", Version: "v1.2.3", Module: "example.com/app", GoVersion: "go1.27.1", Platform: "linux/amd64"}
	if got, want := info.String(), "app v1.2.3 (example.com/app, go1.27.1, linux/amd64)"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
	info.Module = ""
	if got := info.String(); !strings.Contains(got, "unknown module") {
		t.Fatalf("String() = %q, want it to name the unknown module", got)
	}
}
