package cmd

import (
	"log"
	"os"

	"github.com/fastly/agent-autoupdate/cmd/jobs"
)

func initLogging(logFilePath string) (*os.File, error) {

	file, err := os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return nil, err
	}

	log.SetOutput(file)

	// Optional: Set log flags for timestamp
	log.SetFlags(log.LstdFlags)

	return file, nil
}

// Start is called by main.main()
func Start(version string, sha string) {
	logFile, err := initLogging("autoupdate.log")
	if err != nil {
		log.Printf("Error setting up logging:%v\n", err)
	}
	defer logFile.Close()

	log.Printf("Version: %s, BuildSha: %s\n", version, sha)
	jobs.CheckForAgentUpdates()
}
