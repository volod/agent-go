// Package cli owns the command-line contract: argument parsing, the logger,
// the standard streams and exit codes. It passes validated input to domain
// packages and maps their errors to exit codes; no other product package
// writes to stdout or exits the process.
package cli

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"

	"github.com/volod/agent-go/internal/buildinfo"
)

// Exit codes are part of the command-line contract in the specification.
const (
	ExitOK          = 0
	ExitFailure     = 1
	ExitUsage       = 2
	ExitInterrupted = 130
)

// errUsage marks invalid arguments. Run reports errors that wrap it with ExitUsage.
var errUsage = errors.New("usage")

// command is one subcommand. run receives the arguments after the command name;
// a command with flags parses them with its own flag.FlagSet and wraps
// argument errors with errUsage.
type command struct {
	name    string
	summary string
	run     func(ctx context.Context, args []string, stdout io.Writer, log *slog.Logger) error
}

// commands is the command table in help order.
var commands = []command{
	{name: "version", summary: "print the build identity", run: runVersion},
}

// Run executes one command line and returns the process exit code. Canceling
// ctx asks the running command to stop; Run then returns ExitInterrupted.
func Run(ctx context.Context, args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet(buildinfo.Name, flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.Usage = func() { printUsage(stderr) }
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return ExitOK
		}
		return ExitUsage // flag has printed the error and the usage
	}

	name := flags.Arg(0)
	switch name {
	case "":
		printUsage(stderr)
		return ExitUsage
	case "help":
		printUsage(stdout)
		return ExitOK
	}
	for _, c := range commands {
		if c.name == name {
			log := slog.New(slog.NewTextHandler(stderr, nil))
			return exitCode(stderr, c.run(ctx, flags.Args()[1:], stdout, log))
		}
	}
	fmt.Fprintf(stderr, "%s: unknown command %q; run '%s help'\n", buildinfo.Name, name, buildinfo.Name)
	return ExitUsage
}

// exitCode reports err on stderr and maps it to the exit code of the contract.
func exitCode(stderr io.Writer, err error) int {
	switch {
	case err == nil:
		return ExitOK
	case errors.Is(err, context.Canceled):
		fmt.Fprintf(stderr, "%s: interrupted\n", buildinfo.Name)
		return ExitInterrupted
	}
	fmt.Fprintf(stderr, "%s: %v\n", buildinfo.Name, err)
	if errors.Is(err, errUsage) {
		return ExitUsage
	}
	return ExitFailure
}

func printUsage(w io.Writer) {
	fmt.Fprintf(w, "Usage: %s <command> [arguments]\n\nCommands:\n", buildinfo.Name)
	fmt.Fprintf(w, "  %-10s %s\n", "help", "show this help")
	for _, c := range commands {
		fmt.Fprintf(w, "  %-10s %s\n", c.name, c.summary)
	}
}

func runVersion(_ context.Context, args []string, stdout io.Writer, _ *slog.Logger) error {
	if len(args) > 0 {
		return fmt.Errorf("%w: version takes no arguments", errUsage)
	}
	_, err := fmt.Fprintln(stdout, buildinfo.Read())
	return err
}
