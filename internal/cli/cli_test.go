package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/volod/agent-go/internal/buildinfo"
)

func TestRun(t *testing.T) {
	cases := []struct {
		name       string
		args       []string
		wantCode   int
		wantStdout string // substring; "" requires empty stdout
		wantStderr string // substring
	}{
		{"version", []string{"version"}, ExitOK, buildinfo.Name + " ", ""},
		{"help command", []string{"help"}, ExitOK, "version", ""},
		{"help flag", []string{"-h"}, ExitOK, "", "Usage:"},
		{"no command", nil, ExitUsage, "", "Usage:"},
		{"unknown command", []string{"nope"}, ExitUsage, "", `unknown command "nope"`},
		{"unknown flag", []string{"-nope"}, ExitUsage, "", "flag provided but not defined"},
		{"extra argument", []string{"version", "extra"}, ExitUsage, "", "takes no arguments"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			code := Run(t.Context(), tc.args, &stdout, &stderr)
			if code != tc.wantCode {
				t.Errorf("exit code = %d, want %d; stderr:\n%s", code, tc.wantCode, stderr.String())
			}
			if tc.wantStdout == "" && stdout.Len() > 0 {
				t.Errorf("stdout = %q, want empty: diagnostics belong on stderr", stdout.String())
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

func TestExitCode(t *testing.T) {
	cases := []struct {
		name       string
		err        error
		wantCode   int
		wantStderr string
	}{
		{"success", nil, ExitOK, ""},
		{"failure", errors.New("boom"), ExitFailure, buildinfo.Name + ": boom"},
		{"usage", fmt.Errorf("%w: bad flag", errUsage), ExitUsage, "bad flag"},
		{"interrupted", fmt.Errorf("copy: %w", context.Canceled), ExitInterrupted, "interrupted"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var stderr bytes.Buffer
			if got := exitCode(&stderr, tc.err); got != tc.wantCode {
				t.Errorf("exitCode(%v) = %d, want %d", tc.err, got, tc.wantCode)
			}
			if tc.wantStderr == "" && stderr.Len() > 0 {
				t.Errorf("stderr = %q, want empty", stderr.String())
			}
			if !strings.Contains(stderr.String(), tc.wantStderr) {
				t.Errorf("stderr = %q, want it to contain %q", stderr.String(), tc.wantStderr)
			}
		})
	}
}
