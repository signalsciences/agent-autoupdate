package main

import (
	_ "embed"
	"runtime/debug"
	"strings"

	"github.com/signalsciences/agent-autoupdate/cmd"
)

var (
	//go:embed VERSION
	version  string
	Version  = strings.TrimSpace(version)
	BuildSHA = "development"
)

func main() {
	cmd.Start(Version, BuildSHA)
}

// init reads the embedded binary version information at start up.
func init() {
	if info, ok := debug.ReadBuildInfo(); ok {
		for _, s := range info.Settings {
			if "vcs.revision" == s.Key {
				BuildSHA = s.Value
				break
			}
		}
	}
}
