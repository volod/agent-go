package main

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"runtime/debug"
	"slices"
	"strings"
	"testing"
	"time"
)

var commitTime = time.Date(2026, 9, 18, 4, 54, 3, 0, time.UTC)

// member is what a test checks for one archive entry.
type member struct {
	name string
	mode fs.FileMode
	data string
	time time.Time
}

// fixture writes fake binaries and documents into a temporary directory and
// returns a config for them with a reader that reports version for each binary.
func fixture(t *testing.T, version string) (Config, func(string) (stamp, error)) {
	t.Helper()
	dir := t.TempDir()
	files := map[string]string{
		"bin/app-linux-amd64":       "linux binary",
		"bin/app-windows-amd64.exe": "windows binary",
		"README.md":                 "readme",
		"LICENSE":                   "license",
	}
	for name, data := range files {
		path := filepath.Join(dir, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	cfg := Config{
		App:       "app",
		BinDir:    filepath.Join(dir, "bin"),
		OutDir:    filepath.Join(dir, "dist"),
		Files:     []string{filepath.Join(dir, "README.md"), filepath.Join(dir, "LICENSE")},
		Platforms: []string{"linux/amd64", "windows/amd64"},
	}
	read := func(path string) (stamp, error) {
		if _, err := os.Stat(path); err != nil {
			return stamp{}, err
		}
		return stamp{version: version, time: commitTime}, nil
	}
	return cfg, read
}

func TestPackage(t *testing.T) {
	cfg, read := fixture(t, "v1.2.3")
	names, err := Package(cfg, read)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"app-v1.2.3-linux-amd64.tar.gz", "app-v1.2.3-windows-amd64.zip"}
	if !slices.Equal(names, want) {
		t.Fatalf("archives = %q, want %q", names, want)
	}

	stem := "app-v1.2.3-linux-amd64/"
	assertMembers(t, readTarGz(t, filepath.Join(cfg.OutDir, names[0])), []member{
		{stem, fs.ModeDir | 0o755, "", commitTime},
		{stem + "app", 0o755, "linux binary", commitTime},
		{stem + "README.md", 0o644, "readme", commitTime},
		{stem + "LICENSE", 0o644, "license", commitTime},
	})
	stem = "app-v1.2.3-windows-amd64/"
	assertMembers(t, readZip(t, filepath.Join(cfg.OutDir, names[1])), []member{
		{stem, fs.ModeDir | 0o755, "", commitTime},
		{stem + "app.exe", 0o755, "windows binary", commitTime},
		{stem + "README.md", 0o644, "readme", commitTime},
		{stem + "LICENSE", 0o644, "license", commitTime},
	})

	var sums strings.Builder
	for _, name := range names { // already in name order
		data, err := os.ReadFile(filepath.Join(cfg.OutDir, name))
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&sums, "%x  %s\n", sha256.Sum256(data), name)
	}
	if got := readFile(t, filepath.Join(cfg.OutDir, sumsFile)); got != sums.String() {
		t.Fatalf("%s =\n%s\nwant\n%s", sumsFile, got, sums.String())
	}
}

func TestPackageIsReproducible(t *testing.T) {
	cfg, read := fixture(t, "v1.2.3")
	if _, err := Package(cfg, read); err != nil {
		t.Fatal(err)
	}
	first := readFile(t, filepath.Join(cfg.OutDir, sumsFile))
	cfg.OutDir += "-again"
	if _, err := Package(cfg, read); err != nil {
		t.Fatal(err)
	}
	if again := readFile(t, filepath.Join(cfg.OutDir, sumsFile)); again != first {
		t.Fatalf("checksums differ between runs:\n%s\n%s", first, again)
	}
}

func TestPackageRejects(t *testing.T) {
	cases := []struct {
		name   string
		change func(cfg *Config, read *func(string) (stamp, error))
		want   string
	}{
		{"malformed platform", func(cfg *Config, _ *func(string) (stamp, error)) {
			cfg.Platforms = []string{"linux"}
		}, `platform "linux": want GOOS/GOARCH`},
		{"missing binary", func(cfg *Config, _ *func(string) (stamp, error)) {
			cfg.Platforms = append(cfg.Platforms, "darwin/arm64")
		}, "app-darwin-arm64: "},
		{"version other than required", func(cfg *Config, _ *func(string) (stamp, error)) {
			cfg.Version = "v2.0.0"
		}, "reports version v1.2.3, want v2.0.0"},
		{"binaries from different builds", func(_ *Config, read *func(string) (stamp, error)) {
			orig := *read
			*read = func(path string) (stamp, error) {
				s, err := orig(path)
				if strings.HasSuffix(path, ".exe") {
					s.version = "v1.2.4"
				}
				return s, err
			}
		}, "rebuild them together"},
		{"missing document", func(cfg *Config, _ *func(string) (stamp, error)) {
			cfg.Files = append(cfg.Files, filepath.Join(cfg.BinDir, "NOTICE"))
		}, "NOTICE"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			cfg, read := fixture(t, "v1.2.3")
			tc.change(&cfg, &read)
			_, err := Package(cfg, read)
			if err == nil || !strings.Contains(err.Error(), tc.want) {
				t.Fatalf("Package() error = %v, want it to contain %q", err, tc.want)
			}
		})
	}
}

func TestStampOf(t *testing.T) {
	build := func(version string, settings ...debug.BuildSetting) *debug.BuildInfo {
		return &debug.BuildInfo{Main: debug.Module{Version: version}, Settings: settings}
	}
	vcsTime := func(v string) debug.BuildSetting { return debug.BuildSetting{Key: "vcs.time", Value: v} }
	cases := []struct {
		name     string
		bi       *debug.BuildInfo
		wantErr  bool
		wantTime time.Time
	}{
		{"tagged build", build("v1.2.3", vcsTime("2026-09-18T04:54:03Z")), false, commitTime},
		{"no commit time", build("v1.2.3"), false, zipEpoch},
		{"commit time before 1980", build("v1.2.3", vcsTime("1970-01-01T00:00:00Z")), false, zipEpoch},
		{"go run build", build("(devel)"), true, time.Time{}},
		{"no version", build(""), true, time.Time{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			s, err := stampOf(tc.bi)
			if (err != nil) != tc.wantErr {
				t.Fatalf("stampOf() error = %v, want error %v", err, tc.wantErr)
			}
			if !tc.wantErr && (s.version != tc.bi.Main.Version || !s.time.Equal(tc.wantTime)) {
				t.Fatalf("stampOf() = %+v, want version %s and time %s", s, tc.bi.Main.Version, tc.wantTime)
			}
		})
	}
}

func TestReadStampOfTestBinary(t *testing.T) {
	// go test builds without a module version, which dist must refuse.
	exe, err := os.Executable()
	if err != nil {
		t.Fatal(err)
	}
	if _, err := readStamp(exe); err == nil || !strings.Contains(err.Error(), "no module version") {
		t.Fatalf("readStamp(test binary) error = %v, want no module version", err)
	}
}

func TestRun(t *testing.T) {
	cfg, _ := fixture(t, "v1.2.3")
	cases := []struct {
		name     string
		args     []string
		wantCode int
	}{
		{"no app", []string{"linux/amd64"}, 2},
		{"no platform", []string{"-app", "app"}, 2},
		{"unknown flag", []string{"-nope"}, 2},
		// The fixture binaries are not Go binaries, so reading their build info fails.
		{"unreadable binary", []string{"-app", "app", "-bin", cfg.BinDir, "-out", cfg.OutDir, "linux/amd64"}, 1},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			if got := run(tc.args, &stdout, &stderr); got != tc.wantCode {
				t.Fatalf("exit code = %d, want %d; stderr:\n%s", got, tc.wantCode, stderr.String())
			}
		})
	}
}

func assertMembers(t *testing.T, got, want []member) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("archive members = %+v, want %+v", got, want)
	}
	for i := range want {
		g, w := got[i], want[i]
		if g.name != w.name || g.mode != w.mode || g.data != w.data || !g.time.Equal(w.time) {
			t.Errorf("member %d = %+v, want %+v", i, g, w)
		}
	}
}

func readTarGz(t *testing.T, path string) []member {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		t.Fatal(err)
	}
	var members []member
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return members
		}
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(tr)
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, member{hdr.Name, hdr.FileInfo().Mode(), string(data), hdr.ModTime})
	}
}

func readZip(t *testing.T, path string) []member {
	t.Helper()
	zr, err := zip.OpenReader(path)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	var members []member
	for _, f := range zr.File {
		rc, err := f.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			t.Fatal(err)
		}
		members = append(members, member{f.Name, f.Mode(), string(data), f.Modified})
	}
	return members
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
