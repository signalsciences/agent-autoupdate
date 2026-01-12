package jobs

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/mod/semver"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

const (
	DL_DOWNLOAD_SITE = `https://dl.security.fastly.com/sigsci-agent`

	// Windows Registry Key Names
	REGISTRY_KEY_DISPLAY_NAME          = `DisplayName`
	REGISTRY_KEY_PUBLISHER             = `Publisher`
	REGISTRY_KEY_DISPLAY_VERSION       = `DisplayVersion`
	REGISTRY_KEY_UNINSTALLSTR          = `UninstallString`
	REGISTRY_KEY_ENVIRONMENT           = `Environment`
	REGISTRY_KEY_PATH_UNINSTALL        = `SOFTWARE\Microsoft\Windows\CurrentVersion\Uninstall`
	REGISTRY_KEY_PATH_CURRENT_SERVICES = `SYSTEM\CurrentControlSet\Services\sigsci-agent`

	// Windows Registry Value Names
	REGISTRY_VALUE_DISPLAY_NAME = `Signal Sciences Agent`
	REGISTRY_VALUE_PUBLISHER    = `Signal Sciences`

	WIN_SERVICE_NAME = `sigsci-agent`
	STOP_SERVICE     = 1
	START_SERVICE    = 4
)

type agentInfo struct {
	version     string // The current agent version
	productCode string // The GUID given by Windows when the agent is installed
}

// CheckForAgentUpdates is the main entry point for the agent update job.
// The idea being we'll check the windows registry to see if a previous agent is installed,
// then we'll compare versions.
func CheckForAgentUpdates() {

	agent := &agentInfo{}
	if err := agent.getInstalledInfo(); err != nil {
		log.Printf("Agent is not installed: %v\n", err)
		return
	}

	if latest_version, found, err := agent.compareAgainstLatest(); found {
		if err := installNewVersion(latest_version); err != nil {
			log.Printf("installation error: %v\n", err)
		} else {
			log.Printf("installation of version %s complete\n", latest_version)
		}
	} else {
		log.Printf("%v", err)
	}
}

// controlService starts/stops the running sigsci-agent windows service.
func controlService(userCmd uint32) error {
	// Connect to the Windows service manager
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("Connecting to service manager failed: %w\n", err)
	}
	defer func() {
		err := m.Disconnect()
		if err != nil {
			log.Printf("Disconnecting from service manager: %v\n", err)
		}
	}()

	service, err := m.OpenService(WIN_SERVICE_NAME)
	if err != nil {
		return fmt.Errorf("Could not access agent service: %w\n", err)
	}
	defer service.Close()

	status, _ := service.Query()
	// was the service already stopped or started?
	if status.State == svc.Stopped &&
		userCmd == STOP_SERVICE {
		log.Println("Service is already stopped")
		return nil
	}

	if status.State == svc.Running &&
		userCmd == START_SERVICE {
		log.Println("Service is already running")
		return nil
	}

	wait_count := 0
	wait_max := 5
	if userCmd == STOP_SERVICE {
		// Stop the service
		_, err = service.Control(svc.Stop)
		if err != nil {
			// log but don't exit
			log.Printf("Could not stop service: %v\n", err)
		}

		// We only break if the status is Stopped or it times out (error)
		for {
			status, errQuery := service.Query()
			if errQuery != nil {
				log.Printf("Agent service could not be queried(stop), status: %v\n", errQuery)
			}
			if status.State == svc.Stopped {
				err = nil
				break
			}
			if wait_count == wait_max {
				log.Println("Agent stop service max wait reached")
				err = exec.ErrWaitDelay
				break
			}
			time.Sleep(500 * time.Microsecond)
			wait_count++
		}

		if err != nil {
			return err
		}
	} else {
		// Start the service
		err = service.Start()
		if err != nil {
			log.Printf("Could not start agent service: %v\n", err)
		}

		// We only break if the status is Running or it times out (error)
		for {
			status, errQuery := service.Query()
			if errQuery != nil {
				log.Printf("Agent service could not be queried(start), status: %v\n", errQuery)
			}
			if status.State == svc.Running {
				err = nil
				break
			}
			if wait_count == wait_max {
				log.Println("Agent start service max wait reached")
				err = exec.ErrWaitDelay
				break
			}
			time.Sleep(500 * time.Microsecond)
			wait_count++
		}

		if err != nil {
			return err
		}

	}
	return nil
}

// fileExists checks if the path physically exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// installNewVersion will download the latest agent and attempt to install it.
func installNewVersion(version_to_install string) (reterr error) {

	agentVersionName := "sigsci-agent_" + version_to_install + ".msi"
	resp, err := http.Get(DL_DOWNLOAD_SITE + "/" + version_to_install + "/windows/" + agentVersionName)
	if err != nil {
		return fmt.Errorf("net request for version %s of agent, err: %w\n", version_to_install, err)
	}
	defer resp.Body.Close()

	// Process the response
	if resp.StatusCode == http.StatusOK {

		tempDir := os.TempDir()
		msiPath := filepath.Join(tempDir, agentVersionName)

		tempFile, err := os.Create(msiPath)
		if err != nil {
			return err
		}
		defer tempFile.Close()

		if !fileExists(msiPath) {
			return fmt.Errorf("Temporary file path does not exist")
		}
		if _, err = io.Copy(tempFile, resp.Body); err != nil {
			return fmt.Errorf("net I/O failure for %s of agent: %w\n", version_to_install, err)
		}

		_ = tempFile.Sync()

		// Stop the Service, then install.
		err = controlService(STOP_SERVICE)
		if err != nil {
			return err
		}

		msiexeclogName := "sigsci-agent_" + version_to_install + "_msiexec.log"
		msiexeclogPath := filepath.Join(tempDir, msiexeclogName)
		log.Printf("%s\n", msiexeclogPath)
		psCommand := fmt.Sprintf(`Start-Process msiexec.exe -ArgumentList '/i', '"%s"', '/quiet', '/norestart', '/l*vx!', '"%s"' -Verb RunAs`, msiPath, msiexeclogPath)
		log.Printf("%s\n", psCommand)
		cmd := exec.Command("powershell", "-Command", psCommand)

		output, retErr := cmd.CombinedOutput()
		if retErr != nil {

			log.Printf("Failed to install MSI: %v\nOutput: %s\n", retErr, string(output))
			// Don't leave the existing service in a stalled state; force start
			_ = controlService(START_SERVICE)
		}

		// delete the MSI file
		_ = os.Remove(msiPath)

		return retErr

	} else {
		log.Printf("Response code for version %s agent is not ok: %v\n", version_to_install, resp.StatusCode)
	}
	return exec.ErrNotFound

}

// getInstalledInfo checks to see if the agent is installed within the registry and returns the product code
// and any other info required.
func (p *agentInfo) getInstalledInfo() error {

	key, err := registry.OpenKey(registry.LOCAL_MACHINE, REGISTRY_KEY_PATH_UNINSTALL, registry.READ)
	if err != nil {
		return fmt.Errorf("Failed to open registry key: %w\n", err)
	}
	defer key.Close()

	subKeys, err := key.ReadSubKeyNames(-1)

	if err != nil {
		return fmt.Errorf("Failed to read subkeys: %w\n", err)
	}

	for _, subKey := range subKeys {

		subKeyPath := REGISTRY_KEY_PATH_UNINSTALL + `\` + subKey
		//fmt.Println(" subKeyPath: ", subKeyPath, " subKey:", subKey)
		inodeKey, err := registry.OpenKey(registry.LOCAL_MACHINE, subKeyPath, registry.READ)
		if err != nil {
			continue
		}
		defer inodeKey.Close()

		displayName, _, err := inodeKey.GetStringValue(REGISTRY_KEY_DISPLAY_NAME)
		if err == nil {
			if displayName == REGISTRY_VALUE_DISPLAY_NAME {

				publisher, _, err := inodeKey.GetStringValue(REGISTRY_KEY_PUBLISHER)
				if err == nil {
					if publisher == REGISTRY_VALUE_PUBLISHER {

						version, _, err := inodeKey.GetStringValue(REGISTRY_KEY_DISPLAY_VERSION)
						if err == nil {
							p.productCode = subKey
							p.version = version
							return nil
						}
					}
				}
			}
		}
	}
	return exec.ErrNotFound
}

// compareAgainstLatest checks the installed version against the latest agent version online.
func (p *agentInfo) compareAgainstLatest() (string, bool, error) {
	resp, err := http.Get(DL_DOWNLOAD_SITE + "/VERSION")
	if err != nil {
		return "", false, fmt.Errorf("Failed to get latest VERSION: %w\n", err)
	}
	defer resp.Body.Close()

	// Process the response (e.g., compare version numbers)
	if resp.StatusCode == http.StatusOK {

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return "", false, fmt.Errorf("Failed to read VERSION body %w\n", err)
		}
		latest_version := strings.TrimSpace(string(body))
		semver_installed := "v" + p.version
		semvar_latest := "v" + latest_version
		//log.Printf("latest version: %v, installed version: %v\n", semvar_latest, semver_installed)
		if semver.Compare(semvar_latest, semver_installed) >= 1 {
			return latest_version, true, nil
		}
		return "", false, fmt.Errorf("No new agents found")
	}
	return "", false, fmt.Errorf("VERSION status code is not ok: %d\n", resp.StatusCode)
}
