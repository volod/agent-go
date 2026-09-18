// Command dist packages cross-built binaries into release archives:
//
//	dist -app NAME [-bin DIR] [-out DIR] [-version V] [-files "README.md LICENSE"] GOOS/GOARCH...
//
// For each platform it reads DIR/NAME-GOOS-GOARCH[.exe] (the names `make
// build-all` writes), takes the version stamped into the binary, and writes
// NAME-VERSION-GOOS-GOARCH.tar.gz (.zip for Windows) plus SHA256SUMS. The
// archives are reproducible: their bytes depend only on the inputs and the
// commit time. It runs from the repository root and is never shipped.
package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr))
}

func run(args []string, stdout, stderr io.Writer) int {
	flags := flag.NewFlagSet("dist", flag.ContinueOnError)
	flags.SetOutput(stderr)
	var cfg Config
	flags.StringVar(&cfg.App, "app", "", "command `name`; binaries are DIR/NAME-GOOS-GOARCH[.exe]")
	flags.StringVar(&cfg.BinDir, "bin", "bin", "`dir` holding the cross-built binaries")
	flags.StringVar(&cfg.OutDir, "out", "dist", "output `dir`")
	flags.StringVar(&cfg.Version, "version", "", "fail unless every binary reports this `version`")
	files := flags.String("files", "README.md LICENSE", "space-separated `paths` added to every archive")
	if err := flags.Parse(args); err != nil {
		return 2
	}
	cfg.Files, cfg.Platforms = strings.Fields(*files), flags.Args()
	if cfg.App == "" || len(cfg.Platforms) == 0 {
		fmt.Fprintln(stderr, "usage: dist -app NAME [flags] GOOS/GOARCH...")
		return 2
	}
	names, err := Package(cfg, readStamp)
	if err != nil {
		fmt.Fprintln(stderr, "dist:", err)
		return 1
	}
	for _, n := range append(names, sumsFile) {
		fmt.Fprintln(stdout, filepath.Join(cfg.OutDir, n))
	}
	return 0
}
