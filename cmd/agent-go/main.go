// Command agent-go is the starter executable of the template. main only wires
// the process: a context canceled by SIGINT or SIGTERM, the standard streams
// and the exit code. Behavior lives in internal packages.
package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/volod/agent-go/internal/cli"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	code := cli.Run(ctx, os.Args[1:], os.Stdout, os.Stderr)
	stop()
	os.Exit(code)
}
