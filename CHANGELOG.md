# agent-autoupdate Release Notes

# 1.0.2 2026-01-12

* build/release workflow fixes for v1.0.1

# 1.0.1 2026-01-12

* Properly quote windows path args to msiexec
* add embedded version information to binary

# 1.0.0 2025-12-15

* Log msiexec actions to file in Windows temp.
* Log stdout/stderr to console by default; installer actions are captured in the Event Viewer when run from the Task Scheduler.
* Bump Version since we're out of Beta.

# 0.2.2 2025-03-24

* Upgrade to Go 1.23
* Used Github Action to run golangci-lint and upgraded to 1.64
* Built with Windows runner.

# 0.2.1 2025-03-05

* Prepared for open sourcing

# 0.2.0 2024-10-31

* Initial release
