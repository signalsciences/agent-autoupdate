# agent-autoupdate utility
The windows auto-update utility for the [Fastly NGWAF](https://docs.fastly.com/products/fastly-next-gen-waf) works in conjunction with the Windows Task Scheduler.  The Task Scheduler allows one to automate tasks such as running programs, scripts at specified times or events.

## Limitations and considerations

-   Currently the agent auto-update utility assumes prior Fastly agents have been installed using the MSI installer.

## Working with the agent auto-update utility

- The agent auto-update utility does not depend on the Fastly agent to be installed.  The binary can exit independently until a Fastly agent is detected and installed on the machine.

### Install the agent auto-update utility.

-   Installing the msi will create entries in the following folder, `C:\Program Files\Signal Sciences\AgentAutoUpdate`, similar to other Fastly windows modules.

### Steps to Use Windows Task Scheduler

1. Open Task Scheduler
    - Search for Task Scheduler or type `taskschd.msc` in a Run dialog and press Enter. This opens the Task Scheduler window

2. Create a New Task
    - In the Task Scheduler window, click on Action in the top menu, then select Create Task.  Alternatively, in the Actions pane on the right, click Create Task.

3. Define Settings
    - In the General tab:
        - Name: Give the task a name, e.g., "Fastly Agent Update".
        - Security Options:
            Run whether the user is logged on or not (for background tasks).

4. Set a Trigger
    - In the Triggers tab, create a New trigger.
        - Define when the task should start, e.g., On a monthly schedule.
        - Select all Months and select specific "Days" or "On" the first Wednesday, e.g., "On Third Thursday" of the month, which is recommended.
    - Click Ok

5. Define the Action
    - In the Actions tab, create a new Action.
        - Under Action, Start a program.
        - Browse to the agent auto-update installation path aforementioned above and select the binary, `agent-autoupdate.exe`

6.  Save the Task
    - After configuring the task, Click Ok to save it.
    - You will be prompted to enter the password for the selected user account.

7. Testing the task
    - To test if the task is working, right-click on the task in the Task Scheduler Library and Select Run.  Be sure to enable Task History: In the Task Scheduler, enable the history for the task to view detailed logs of when the task ran and any issues that might have occurred. Right-click the task in the Task Scheduler Library and select Enable All Tasks History. 

8. Subsequent releases of the agent auto-update utility do not require recreating a new task unless otherwise explicitly stated in the CHANGELOG.md

### Affirmation

-   When new Fastly agents are downloaded and installed, **MsiInstaller** source entries in the Windows Event Viewer will be made such as the following:

        `Windows Installer installed the product. Product Name: Signal Sciences Agent. Product Version: 4.59.1. Product Language: 1033. Manufacturer: Signal Sciences. Installation success or error status: 0.`

### Troubleshooting

-   If the scheduled task should fail, the agent auto-update utility can be run manually by navigating to the installed directory and running the binary in an Administrative Command Prompt.  A log file capturing error diagnostics will be created.

-   Enable Task History and Check Event Viewer: If a task fails or doesn't run as expected, you can also check the Windows Event Viewer for relevant logs. Look in the TaskScheduler log under Windows Logs → System
