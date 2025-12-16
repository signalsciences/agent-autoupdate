package cmd

import (
	"log"

	"github.com/signalsciences/agent-autoupdate/cmd/jobs"
)

// Start is called by main.main()
func Start(version string, sha string) {
	log.Printf("Version: %s, BuildSha: %s\n", version, sha)
	jobs.CheckForAgentUpdates()
}
