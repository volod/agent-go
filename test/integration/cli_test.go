// Package integration_test drives the built command as a black box: it checks
// the command-line contract through a real process, its streams and its exit
// code.
package integration_test

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestCommandLineContract(t *testing.T) {
	bin := buildCommand(t)
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string
		wantStderr string
	}{
		{"version", []string{"version"}, 0, "agent-go ", ""},
		{"version names the module", []string{"version"}, 0, "github.com/volod/agent-go", ""},
		{"help", []string{"help"}, 0, "version", ""},
		{"unknown command", []string{"nope"}, 2, "", "unknown command"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			cmd := exec.CommandContext(t.Context(), bin, tc.args...)
			cmd.Stdout, cmd.Stderr = &stdout, &stderr
			var exitErr *exec.ExitError
			if err := cmd.Run(); err != nil && !errors.As(err, &exitErr) {
				t.Fatalf("run %s: %v", bin, err)
			}
			if got := cmd.ProcessState.ExitCode(); got != tc.wantCode {
				t.Errorf("exit code = %d, want %d; stderr:\n%s", got, tc.wantCode, stderr.String())
			}
			if !strings.Contains(stdout.String(), tc.wantStdout) {
				t.Errorf("stdout = %q, want it to contain %q", stdout.String(), tc.wantStdout)
			}
			if !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}

// buildCommand builds cmd/agent-go into a temporary directory the way `make
// build` does and returns the executable path.
func buildCommand(t *testing.T) string {
	t.Helper()
	bin := filepath.Join(t.TempDir(), "agent-go")
	if runtime.GOOS == "windows" {
		bin += ".exe"
	}
	build := exec.CommandContext(t.Context(), "go", "build", "-trimpath", "-o", bin, "../../cmd/agent-go")
	build.Env = append(os.Environ(), "CGO_ENABLED=0")
	if out, err := build.CombinedOutput(); err != nil {
		t.Fatalf("go build: %v\n%s", err, out)
	}
	return bin
}
