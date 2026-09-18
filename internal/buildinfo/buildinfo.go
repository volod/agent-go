// Package buildinfo reports the identity of the running binary. The version
// comes from the Go toolchain's build stamping: a tagged commit reports its
// tag, an untagged `go build` a pseudo-version, and `go run` or `go test`
// report [DevelVersion].
package buildinfo

import (
	"fmt"
	"runtime"
	"runtime/debug"
)

// Name is the command name. Personalizing the template changes it.
const Name = "agent-go"

// DevelVersion is reported when the build carries no module version.
const DevelVersion = "devel"

// Info identifies one build of the command.
type Info struct {
	Name      string
	Module    string // main module path; empty when the build has no module information
	Version   string // module version, or DevelVersion
	GoVersion string
	Platform  string // GOOS/GOARCH
}

// Read returns the identity of the running binary.
func Read() Info {
	bi, _ := debug.ReadBuildInfo()
	return fromBuildInfo(bi)
}

// fromBuildInfo is Read without the process dependency; bi may be nil.
func fromBuildInfo(bi *debug.BuildInfo) Info {
	info := Info{
		Name:      Name,
		Version:   DevelVersion,
		GoVersion: runtime.Version(),
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
	if bi == nil {
		return info
	}
	info.Module = bi.Main.Path
	if v := bi.Main.Version; v != "" && v != "(devel)" {
		info.Version = v
	}
	return info
}

// String formats the identity as one line, for example
// "agent-go v0.1.0 (github.com/volod/agent-go, go1.27.1, linux/amd64)".
func (i Info) String() string {
	module := i.Module
	if module == "" {
		module = "unknown module"
	}
	return fmt.Sprintf("%s %s (%s, %s, %s)", i.Name, i.Version, module, i.GoVersion, i.Platform)
}
