# Cabinet isolated WSL runner pool

Cabinet can use one dedicated WSL 2 distribution per self-hosted GitHub Actions
runner. Each member has its own bounded VHDX, Docker daemon, Linux account,
workspace, cleanup timer, GitHub registration, and Windows autostart task.

Repository scope remains the default. Organisation scope can place new runners
in a runner group restricted to selected repositories.

The installer does not change workflow `runs-on` selectors. Merge and provision
the runners first, confirm that they are online, then change selected workflows
in a separately reviewed issue.

## Prerequisites

- Current WSL from `wsl --update`, with WSL 2 available.
- PowerShell 5.1 or newer.
- GitHub CLI authenticated as an account that can administer repository Actions
  runners: `gh auth status`.
- Enough physical space for each requested 60 GB virtual ceiling plus 15 GB of
  Windows headroom. The ceiling is not preallocated, but it is the safe capacity
  used by the install guard.

Do not put a GitHub token or Linux password in shell history. Let the script
prompt securely and let `gh` request short-lived registration tokens.

## Review the plan

From the Cabinet repository root:

```powershell
$password = Read-Host "Linux runner password" -AsSecureString
.\scripts\github-actions-runner\install-wsl-runner.ps1 `
  -LinuxPassword $password `
  -RunnerCount 3 `
  -WhatIf
```

The three-member default plan is:

| Member | WSL distribution | VHD location | Linux user | Unique label |
| --- | --- | --- | --- | --- |
| 1 | `cabinet` | `C:\WSL\cabinet` | `cabinet` | base labels |
| 2 | `cabinet-02` | `C:\WSL\cabinet-02` | `cabinet-02` | `cabinet-runner-02` |
| 3 | `cabinet-03` | `C:\WSL\cabinet-03` | `cabinet-03` | `cabinet-runner-03` |

Use `-InstallLocation D:\WSL\cabinet` to put every member on another drive.

## Create or scale the pool

```powershell
$password = Read-Host "Linux runner password" -AsSecureString
.\scripts\github-actions-runner\install-wsl-runner.ps1 `
  -LinuxPassword $password `
  -RunnerCount 3
```

`RunnerCount` is desired total capacity, not "add this many". Rerunning the same
command skips members that already contain an Actions runner systemd service and
creates only missing members. `gh` requests a fresh single-use registration
token for each member that needs configuration.

The installer adds Docker Engine with `overlay2` and bounded `json-file` logs,
Node.js 22, the Go version in `go.mod`, PowerShell, and official Actions runner
dependencies.

## Inspect runners

```powershell
wsl --list --verbose
wsl -d cabinet -- systemctl status 'actions.runner.*'
wsl -d cabinet -- docker info
wsl -d cabinet -- node --version
wsl -d cabinet -- go version
wsl -d cabinet -- systemctl list-timers cabinet-runner-reaper.timer
Get-ScheduledTask -TaskName "WSL Runner Autostart - cabinet*"
```

In GitHub, check `Settings -> Actions -> Runners`. Every member should have the
base `cabinet-ci,docker,node22,go,wsl` labels; numbered members also have their
unique `cabinet-runner-02` style label.

## Storage controls

At job start, the runner creates an active marker under `/run/cabinet-runner`.
At completion it removes the completed workspace and prunes unused Docker
resources. The recurring reaper also checks the marker and `Runner.Worker`
process before pruning. It bounds journal storage, cleans apt cache, and runs
`fstrim /` at most daily so deleted ext4 blocks can later be reclaimed from the
VHDX.

Inspect usage without deleting anything:

```powershell
wsl -d cabinet -- df -h /
wsl -d cabinet -- docker system df -v
wsl -d cabinet -- sudo du -xhd1 /var /home 2>/dev/null
wsl -d cabinet -- journalctl --disk-usage
```

Run the idle-aware reaper immediately:

```powershell
wsl -d cabinet -- sudo systemctl start cabinet-runner-reaper.service
wsl -d cabinet -- sudo journalctl -u cabinet-runner-reaper.service -n 100 --no-pager
```

## Remove or rebuild one member

Confirm that it is idle, remove it from the repository's GitHub runner settings,
then unregister only the dedicated distribution:

```powershell
wsl --terminate cabinet-03
wsl --unregister cabinet-03
Unregister-ScheduledTask -TaskName "WSL Runner Autostart - cabinet-03" -Confirm:$false
```

`wsl --unregister` permanently deletes that distro and its VHDX. Never use it
against a general-purpose distribution.

To rebuild through the installer, use the matching member parameters with
`-ForceRecreate`. The switch is deliberately destructive and prompts unless
confirmation is suppressed explicitly.

## Create an organisation runner group

GitHub CLI needs organisation runner administration permission (`gh auth
refresh --hostname github.com --scopes admin:org` when needed). This example
creates or reuses `cabinet-wsl` and adds selected repositories without removing
existing access:

```powershell
$password = Read-Host "Linux runner password" -AsSecureString
.\scripts\github-actions-runner\install-wsl-runner.ps1 `
  -LinuxPassword $password `
  -DistroName cabinet-org-02 `
  -InstallLocation C:\WSL\cabinet-org-02 `
  -LinuxUser cabinet-org-02 `
  -RunnerName "$($env:COMPUTERNAME.ToLowerInvariant())-cabinet-org-02" `
  -RunnerScope Organization `
  -Organization collectors-tech `
  -RunnerGroupName cabinet-wsl `
  -RunnerGroupRepositories collectors-tech/another-trusted-repository
```

Cabinet is always included. Public repositories require
`-AllowPublicRepositories`. Existing configured runners are left unchanged
instead of being silently converted to organisation scope.

## Reclaim Windows VHDX space

Pruning files inside Linux does not immediately shrink the Windows VHDX file.
After cleanup:

1. Confirm every runner is idle.
2. Run `fstrim` in each target distro.
3. Stop WSL with `wsl --shutdown`.
4. Use the reviewed `wsl-compact.ps1` procedure from the Windows dev-setup repo
   to compact the now-unmounted VHDX files.

Do not compact a VHDX while its distribution or another process is using it.
