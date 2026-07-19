#Requires -Version 5.1

<#
.SYNOPSIS
    Creates one or more dedicated WSL 2 GitHub Actions runners for Cabinet.

.DESCRIPTION
    Reconciles an isolated runner pool. Every member receives its own bounded
    Ubuntu WSL distribution, Linux account, Docker daemon, work directory,
    GitHub registration, idle-aware cleanup timer, and Windows logon autostart
    task. Existing configured members are left unchanged.

    Passwords and registration tokens cross into WSL through temporary files
    restricted to the current Windows user. When no token is supplied, gh CLI
    requests a fresh repository registration token for every missing member.

.PARAMETER RunnerCount
    Total desired pool size. Member 1 uses the unnumbered names. Additional
    members use -02, -03, and so on. Default: 1.

.PARAMETER VhdSize
    Virtual capacity ceiling for each new distribution. This is not
    preallocated Windows disk usage. Default: 60GB.

.PARAMETER ForceRecreate
    Permanently unregisters an existing targeted distribution before rebuilding
    it. This deletes that member's VHDX and all data inside it.

.EXAMPLE
    $password = Read-Host "Linux runner password" -AsSecureString
    .\install-wsl-runner.ps1 -LinuxPassword $password -WhatIf

.EXAMPLE
    $password = Read-Host "Linux runner password" -AsSecureString
    .\install-wsl-runner.ps1 -LinuxPassword $password -RunnerCount 3
#>

[CmdletBinding(SupportsShouldProcess = $true, ConfirmImpact = "Medium")]
param(
    [Alias("Token")]
    [Security.SecureString]$GitHubToken,

    [string]$GitHubTokenFile,

    [ValidatePattern('^[A-Za-z0-9._-]+$')]
    [string]$DistroName = "cabinet",

    [string]$InstallLocation = "C:\WSL\cabinet",

    [ValidateRange(1, 16)]
    [int]$RunnerCount = 1,

    [ValidatePattern('^[1-9][0-9]*(MB|GB|TB)$')]
    [string]$VhdSize = "60GB",

    [ValidatePattern('^[a-z_][a-z0-9_-]*$')]
    [string]$LinuxUser = "cabinet",

    [Security.SecureString]$LinuxPassword,

    [ValidatePattern('^https://github\.com/[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+/?$')]
    [string]$RepositoryUrl = "https://github.com/collectors-tech/cabinet",

    [ValidatePattern('^[A-Za-z0-9._-]+$')]
    [string]$RunnerName = ("{0}-cabinet-wsl" -f $env:COMPUTERNAME.ToLowerInvariant()),

    [ValidatePattern('^[A-Za-z0-9._-]+(,[A-Za-z0-9._-]+)*$')]
    [string]$RunnerLabels = "cabinet-ci,docker,node22,go,wsl",

    [ValidatePattern('^(latest|[0-9]+(\.[0-9]+){2,3})$')]
    [string]$RunnerVersion = "latest",

    [ValidatePattern('^[0-9]+(\.[0-9]+){2}$')]
    [string]$PowerShellVersion = "7.6.3",

    [ValidatePattern('^[0-9]+(\.[0-9]+){1,2}$')]
    [string]$GoVersion = "",

    [ValidatePattern('^$|^[0-9a-fA-F]{64}$')]
    [string]$RunnerSha256 = "",

    [ValidatePattern('^[1-9][0-9]*(s|m|h)$')]
    [string]$PruneUntil = "1h",

    [ValidateRange(300, 86400)]
    [int]$PruneIntervalSeconds = 1800,

    [switch]$ForceRecreate,
    [switch]$AllowLowDiskSpace,
    [switch]$SkipAutostart
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function Write-Step {
    param([Parameter(Mandatory = $true)][string]$Message)
    Write-Host ("==> {0}" -f $Message) -ForegroundColor Cyan
}

function ConvertFrom-SecureValue {
    param([Parameter(Mandatory = $true)][Security.SecureString]$Value)

    $pointer = [Runtime.InteropServices.Marshal]::SecureStringToBSTR($Value)
    try {
        return [Runtime.InteropServices.Marshal]::PtrToStringBSTR($pointer)
    } finally {
        [Runtime.InteropServices.Marshal]::ZeroFreeBSTR($pointer)
    }
}

function New-RestrictedTemporaryFile {
    param([Parameter(Mandatory = $true)][string]$Content)

    $path = Join-Path ([IO.Path]::GetTempPath()) ("cabinet-runner-{0}.secret" -f [guid]::NewGuid().ToString("N"))
    [IO.File]::WriteAllText($path, $Content, [Text.UTF8Encoding]::new($false))
    try {
        $identity = [Security.Principal.WindowsIdentity]::GetCurrent().User
        $acl = [Security.AccessControl.FileSecurity]::new()
        $acl.SetAccessRuleProtection($true, $false)
        $rule = [Security.AccessControl.FileSystemAccessRule]::new(
            $identity,
            [Security.AccessControl.FileSystemRights]::FullControl,
            [Security.AccessControl.AccessControlType]::Allow
        )
        [void]$acl.AddAccessRule($rule)
        Set-Acl -LiteralPath $path -AclObject $acl
    } catch {
        Remove-Item -LiteralPath $path -Force -ErrorAction SilentlyContinue
        throw "Could not restrict temporary secret file permissions: $($_.Exception.Message)"
    }
    return $path
}

function Invoke-NativeChecked {
    param(
        [Parameter(Mandatory = $true)][string]$FilePath,
        [Parameter(Mandatory = $true)][string[]]$ArgumentList,
        [Parameter(Mandatory = $true)][string]$FailureMessage
    )

    & $FilePath @ArgumentList
    if ($LASTEXITCODE -ne 0) {
        throw ("{0} (exit code {1})" -f $FailureMessage, $LASTEXITCODE)
    }
}

function Get-WslDistributionNames {
    $output = & wsl.exe --list --quiet 2>$null
    if ($LASTEXITCODE -ne 0) {
        throw "Could not list WSL distributions. Run 'wsl --install' first."
    }
    return @($output | ForEach-Object { ([string]$_).Replace("`0", "").Trim() } | Where-Object { $_ })
}

function Test-WslRunnerConfigured {
    param([Parameter(Mandatory = $true)][string]$Distribution)

    & wsl.exe --distribution $Distribution --user root --exec sh -c (
        "systemctl list-unit-files --type=service --no-legend 'actions.runner.*.service' " +
        "| grep -q '^actions\.runner\.'"
    ) *> $null
    return $LASTEXITCODE -eq 0
}

function ConvertTo-WslPath {
    param(
        [Parameter(Mandatory = $true)][string]$WindowsPath,
        [Parameter(Mandatory = $true)][string]$Distribution
    )

    $result = & wsl.exe --distribution $Distribution --user root --exec wslpath -a -u $WindowsPath
    if ($LASTEXITCODE -ne 0) {
        throw "Could not translate a temporary file path for WSL."
    }
    return (([string]($result | Select-Object -First 1)).Replace("`0", "").Trim())
}

function Quote-BashArgument {
    param([Parameter(Mandatory = $true)][AllowEmptyString()][string]$Value)
    $escapedSingleQuote = "'`"'`"'"
    return "'" + $Value.Replace("'", $escapedSingleQuote) + "'"
}

function Invoke-WslBootstrap {
    param(
        [Parameter(Mandatory = $true)][string]$Distribution,
        [Parameter(Mandatory = $true)][string]$Script,
        [Parameter(Mandatory = $true)][string[]]$Arguments,
        [Parameter(Mandatory = $true)][string]$FailureMessage
    )

    $encoded = [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($Script))
    $quotedArguments = ($Arguments | ForEach-Object { Quote-BashArgument $_ }) -join " "
    $command = "printf '%s' '$encoded' | base64 --decode | bash -s -- $quotedArguments"
    & wsl.exe --distribution $Distribution --user root --exec sh -c $command
    if ($LASTEXITCODE -ne 0) {
        throw ("{0} (exit code {1})" -f $FailureMessage, $LASTEXITCODE)
    }
}

function Get-GhRunnerRegistrationToken {
    param([Parameter(Mandatory = $true)][string]$RepoUrl)

    $uri = [Uri]$RepoUrl
    $repoSlug = $uri.AbsolutePath.Trim('/').TrimEnd('/')
    Write-Step "Requesting a fresh runner registration token with GitHub CLI"
    $tokenOutput = & gh.exe api --method POST "repos/$repoSlug/actions/runners/registration-token" --jq .token
    if ($LASTEXITCODE -ne 0) {
        throw "GitHub CLI could not create a runner token for $repoSlug. Confirm gh auth and repository administration access."
    }
    $tokenValue = ([string]($tokenOutput | Select-Object -First 1)).Trim()
    if ([string]::IsNullOrWhiteSpace($tokenValue)) {
        throw "GitHub CLI returned an empty runner registration token."
    }
    return $tokenValue
}

function Get-VhdCeilingBytes {
    param([Parameter(Mandatory = $true)][string]$Size)

    $match = [regex]::Match($Size, '^(?<value>[1-9][0-9]*)(?<unit>MB|GB|TB)$')
    $multiplier = switch ($match.Groups['unit'].Value) {
        'MB' { 1MB }
        'GB' { 1GB }
        'TB' { 1TB }
    }
    return [int64]$match.Groups['value'].Value * [int64]$multiplier
}

function Assert-DriveCapacity {
    param(
        [Parameter(Mandatory = $true)][string]$Location,
        [Parameter(Mandatory = $true)][int]$NewDistributionCount
    )

    if ($NewDistributionCount -eq 0 -or $AllowLowDiskSpace) {
        return
    }
    $fullPath = [IO.Path]::GetFullPath($Location)
    $root = [IO.Path]::GetPathRoot($fullPath)
    $drive = Get-PSDrive -Name $root.TrimEnd('\').TrimEnd(':') -ErrorAction SilentlyContinue
    if ($null -eq $drive) {
        throw "The target drive for '$fullPath' is not available."
    }
    $required = ([int64]$NewDistributionCount * (Get-VhdCeilingBytes $VhdSize)) + 15GB
    if ($drive.Free -lt $required) {
        throw ("Drive {0} has {1:N2} GB free, but {2} new runner VHDs with a {3} ceiling need {4:N2} GB to retain 15 GB of Windows headroom. Choose another -InstallLocation, lower -VhdSize, reduce -RunnerCount, or pass -AllowLowDiskSpace explicitly." -f
            $root, ($drive.Free / 1GB), $NewDistributionCount, $VhdSize, ($required / 1GB))
    }
}

function Register-WslAutostartTask {
    param([Parameter(Mandatory = $true)][string]$Distribution)

    if (-not (Get-Command Register-ScheduledTask -ErrorAction SilentlyContinue)) {
        throw "The Windows ScheduledTasks module is required for WSL autostart."
    }
    $taskName = "WSL Runner Autostart - $Distribution"
    $currentUser = [Security.Principal.WindowsIdentity]::GetCurrent().Name
    $powershellPath = Join-Path $env:SystemRoot "System32\WindowsPowerShell\v1.0\powershell.exe"
    $directory = Join-Path $env:LOCALAPPDATA ("WSLRunnerAutostart\{0}" -f $Distribution)
    $keepalivePath = Join-Path $directory "wsl-runner-keepalive.ps1"
    New-Item -ItemType Directory -Path $directory -Force | Out-Null
    $keepalive = @'
#Requires -Version 5.1
param(
    [Parameter(Mandatory = $true)]
    [ValidatePattern('^[A-Za-z0-9._-]+$')]
    [string]$DistroName
)
$ErrorActionPreference = "Continue"
$wslPath = Join-Path $env:SystemRoot "System32\wsl.exe"
while ($true) {
    & $wslPath --distribution $DistroName --user root --exec /usr/bin/sleep infinity
    Start-Sleep -Seconds 15
}
'@
    [IO.File]::WriteAllText($keepalivePath, $keepalive, [Text.UTF8Encoding]::new($false))

    $existing = Get-ScheduledTask -TaskName $taskName -ErrorAction SilentlyContinue
    if ($existing) {
        if ($existing.State -eq "Running") {
            Stop-ScheduledTask -TaskName $taskName
        }
        Unregister-ScheduledTask -TaskName $taskName -Confirm:$false
    }
    $action = New-ScheduledTaskAction -Execute $powershellPath -Argument (
        '-NoLogo -NoProfile -NonInteractive -ExecutionPolicy Bypass -WindowStyle Hidden -File "{0}" -DistroName "{1}"' -f
            $keepalivePath, $Distribution
    )
    $trigger = New-ScheduledTaskTrigger -AtLogOn -User $currentUser
    $principal = New-ScheduledTaskPrincipal -UserId $currentUser -LogonType Interactive -RunLevel Limited
    $settings = New-ScheduledTaskSettingsSet -AllowStartIfOnBatteries -DontStopIfGoingOnBatteries `
        -ExecutionTimeLimit ([TimeSpan]::Zero) -Hidden -MultipleInstances IgnoreNew `
        -RestartCount 3 -RestartInterval (New-TimeSpan -Minutes 1) -StartWhenAvailable
    Register-ScheduledTask -TaskName $taskName -Action $action -Trigger $trigger `
        -Principal $principal -Settings $settings `
        -Description "Keep the $Distribution WSL GitHub runner running after logon." -Force | Out-Null
    Start-ScheduledTask -TaskName $taskName
}

if (-not (Get-Command wsl.exe -ErrorAction SilentlyContinue)) {
    throw "wsl.exe is required. Install or update WSL first."
}
if ($GitHubToken -and $GitHubTokenFile) {
    throw "Use -GitHubToken or -GitHubTokenFile, not both."
}
if (-not [string]::IsNullOrWhiteSpace($GitHubTokenFile)) {
    $GitHubTokenFile = (Resolve-Path -LiteralPath $GitHubTokenFile -ErrorAction Stop).Path
    if ([string]::IsNullOrWhiteSpace([IO.File]::ReadAllText($GitHubTokenFile))) {
        throw "GitHubTokenFile must not be empty."
    }
}
$useGitHubCli = -not $GitHubToken -and [string]::IsNullOrWhiteSpace($GitHubTokenFile)
if ($useGitHubCli) {
    if (-not (Get-Command gh.exe -ErrorAction SilentlyContinue)) {
        throw "gh.exe is required when no token is supplied."
    }
    & gh.exe auth status --hostname github.com *> $null
    if ($LASTEXITCODE -ne 0) {
        throw "GitHub CLI is not authenticated. Run 'gh auth login' and retry."
    }
}
if ([string]::IsNullOrWhiteSpace($GoVersion)) {
    $goMod = Join-Path (Split-Path -Parent (Split-Path -Parent $PSScriptRoot)) "go.mod"
    $goLine = Get-Content -LiteralPath $goMod | Where-Object { $_ -match '^go\s+[0-9]+\.[0-9]+(\.[0-9]+)?\s*$' } | Select-Object -First 1
    if (-not $goLine) {
        throw "Could not read the Go version from go.mod."
    }
    $GoVersion = ([regex]::Match($goLine, '[0-9]+\.[0-9]+(\.[0-9]+)?')).Value
}
if (-not $LinuxPassword) {
    $LinuxPassword = Read-Host ("Password for WSL user '{0}'" -f $LinuxUser) -AsSecureString
}

# RunnerCount means the total desired pool size, including the existing
# unnumbered member. Recursion with RunnerCount 1 gives every missing member an
# independent distro, registration flow, Docker daemon, cleanup, and autostart.
if ($RunnerCount -gt 1) {
    $existingDistributions = Get-WslDistributionNames
    $poolMembers = @()
    for ($index = 1; $index -le $RunnerCount; $index++) {
        if ($index -eq 1) {
            $memberDistroName = $DistroName
            $memberInstallLocation = $InstallLocation
            $memberLinuxUser = $LinuxUser
            $memberRunnerName = $RunnerName
            $memberRunnerLabels = $RunnerLabels
        } else {
            $memberDistroName = "{0}-{1:D2}" -f $DistroName, $index
            $memberInstallLocation = "{0}-{1:D2}" -f $InstallLocation.TrimEnd('\'), $index
            $memberLinuxUser = "{0}-{1:D2}" -f $LinuxUser, $index
            $memberRunnerName = "{0}-{1:D2}" -f $RunnerName, $index
            $memberRunnerLabels = "{0},cabinet-runner-{1:D2}" -f $RunnerLabels, $index
        }
        if ($memberLinuxUser.Length -gt 32) {
            throw "Generated Linux user '$memberLinuxUser' exceeds the 32-character account-name limit."
        }
        $memberExists = $existingDistributions -contains $memberDistroName
        $memberConfigured = $false
        if ($memberExists -and -not $ForceRecreate) {
            $memberConfigured = Test-WslRunnerConfigured -Distribution $memberDistroName
        }
        $poolMembers += [pscustomobject]@{
            Index = $index
            DistroName = $memberDistroName
            InstallLocation = $memberInstallLocation
            LinuxUser = $memberLinuxUser
            RunnerName = $memberRunnerName
            RunnerLabels = $memberRunnerLabels
            Exists = $memberExists
            Configured = $memberConfigured
        }
    }
    $newCount = @($poolMembers | Where-Object { -not $_.Exists -or $ForceRecreate }).Count
    $membersToConfigure = @($poolMembers | Where-Object { -not $_.Configured -or $ForceRecreate })
    if ($membersToConfigure.Count -gt 1 -and -not $useGitHubCli) {
        throw ("{0} pool members require registration, but a supplied GitHub runner registration token can only be used once. Omit -GitHubToken/-GitHubTokenFile so gh creates a fresh token for every member, or configure one runner at a time." -f $membersToConfigure.Count)
    }
    Assert-DriveCapacity -Location $InstallLocation -NewDistributionCount $newCount

    foreach ($member in $poolMembers) {
        if ($member.Configured -and -not $ForceRecreate) {
            Write-Step ("{0} already has a configured runner; leaving it unchanged" -f $member.DistroName)
            continue
        }
        $childParameters = @{
            DistroName = $member.DistroName
            InstallLocation = $member.InstallLocation
            RunnerCount = 1
            VhdSize = $VhdSize
            LinuxUser = $member.LinuxUser
            LinuxPassword = $LinuxPassword
            RepositoryUrl = $RepositoryUrl
            RunnerName = $member.RunnerName
            RunnerLabels = $member.RunnerLabels
            RunnerVersion = $RunnerVersion
            PowerShellVersion = $PowerShellVersion
            GoVersion = $GoVersion
            RunnerSha256 = $RunnerSha256
            PruneUntil = $PruneUntil
            PruneIntervalSeconds = $PruneIntervalSeconds
        }
        if ($GitHubToken) { $childParameters.GitHubToken = $GitHubToken }
        elseif ($GitHubTokenFile) { $childParameters.GitHubTokenFile = $GitHubTokenFile }
        if ($ForceRecreate) { $childParameters.ForceRecreate = $true }
        if ($AllowLowDiskSpace) { $childParameters.AllowLowDiskSpace = $true }
        if ($SkipAutostart) { $childParameters.SkipAutostart = $true }
        if ($WhatIfPreference) { $childParameters.WhatIf = $true }
        if ($PSBoundParameters.ContainsKey('Confirm')) { $childParameters.Confirm = $PSBoundParameters['Confirm'] }
        Write-Step ("Provisioning pool member {0}/{1}: {2}" -f $member.Index, $RunnerCount, $member.DistroName)
        & $PSCommandPath @childParameters
    }
    Write-Host ("Cabinet WSL runner pool plan/configuration complete ({0} members)." -f $RunnerCount) -ForegroundColor Green
    return
}

$distroExists = (Get-WslDistributionNames) -contains $DistroName
$resolvedInstallLocation = [IO.Path]::GetFullPath($InstallLocation)
if ($distroExists -and $ForceRecreate) {
    if ($PSCmdlet.ShouldProcess($DistroName, "Permanently unregister and delete the WSL distribution")) {
        Invoke-NativeChecked wsl.exe @("--unregister", $DistroName) "Could not unregister $DistroName"
        $distroExists = $false
    }
}
Assert-DriveCapacity -Location $resolvedInstallLocation -NewDistributionCount $(if ($distroExists) { 0 } else { 1 })
if (-not $distroExists) {
    if ($PSCmdlet.ShouldProcess($DistroName, "Create Ubuntu WSL distribution at $resolvedInstallLocation with a $VhdSize ceiling")) {
        if (Test-Path -LiteralPath $resolvedInstallLocation) {
            if (@(Get-ChildItem -LiteralPath $resolvedInstallLocation -Force -ErrorAction SilentlyContinue).Count -gt 0) {
                throw "InstallLocation already exists and is not empty: $resolvedInstallLocation"
            }
        }
        Invoke-NativeChecked wsl.exe @(
            "--install", "Ubuntu", "--name", $DistroName, "--location", $resolvedInstallLocation,
            "--vhd-size", $VhdSize, "--no-launch"
        ) "WSL could not create $DistroName"
    } else {
        return
    }
} else {
    Write-Step "Using existing WSL distribution $DistroName"
    if ($WhatIfPreference) {
        Write-Step "WhatIf: would provision the existing distribution and configure its runner"
        return
    }
}

$plainPassword = ConvertFrom-SecureValue $LinuxPassword
$plainToken = ""
$passwordTemp = $null
$tokenTemp = $null
try {
    if ([string]::IsNullOrEmpty($plainPassword)) { throw "The Linux password must not be empty." }
    $passwordTemp = New-RestrictedTemporaryFile $plainPassword
    $plainPassword = ""
    Invoke-NativeChecked wsl.exe @("--distribution", $DistroName, "--user", "root", "--exec", "true") "Could not start $DistroName"
    $passwordWslPath = ConvertTo-WslPath $passwordTemp $DistroName

    $baseBootstrap = @'
set -Eeuo pipefail
linux_user="$1"
password_file="$2"
powershell_version="$3"
go_version="$4"
export DEBIAN_FRONTEND=noninteractive
apt-get update
apt-get install -y --no-install-recommends \
  apt-transport-https build-essential ca-certificates curl git gnupg iproute2 iptables \
  iputils-ping jq libicu-dev libssl-dev openssh-client procps sudo tar uidmap unzip \
  util-linux xz-utils zlib1g
if ! id "${linux_user}" >/dev/null 2>&1; then
  useradd --create-home --home-dir "/home/${linux_user}" --shell /bin/bash "${linux_user}"
fi
password_value="$(tr -d '\r\n' < "${password_file}")"
[[ -n "${password_value}" ]] || { echo "Linux password is empty" >&2; exit 2; }
printf '%s:%s\n' "${linux_user}" "${password_value}" | chpasswd
unset password_value
usermod -aG sudo "${linux_user}"

install -m 0755 -d /etc/apt/keyrings
curl -fsSL https://download.docker.com/linux/ubuntu/gpg -o /etc/apt/keyrings/docker.asc
chmod a+r /etc/apt/keyrings/docker.asc
. /etc/os-release
printf 'deb [arch=%s signed-by=/etc/apt/keyrings/docker.asc] https://download.docker.com/linux/ubuntu %s stable\n' \
  "$(dpkg --print-architecture)" "${VERSION_CODENAME}" > /etc/apt/sources.list.d/docker.list
curl -fsSL https://deb.nodesource.com/gpgkey/nodesource-repo.gpg.key \
  | gpg --dearmor --yes -o /etc/apt/keyrings/nodesource.gpg
printf '%s\n' 'deb [signed-by=/etc/apt/keyrings/nodesource.gpg] https://deb.nodesource.com/node_22.x nodistro main' \
  > /etc/apt/sources.list.d/nodesource.list
apt-get update
apt-get install -y --no-install-recommends \
  containerd.io docker-buildx-plugin docker-ce docker-ce-cli docker-compose-plugin nodejs

case "$(dpkg --print-architecture)" in
  amd64) powershell_arch="amd64"; go_arch="amd64" ;;
  arm64) powershell_arch="arm64"; go_arch="arm64" ;;
  *) echo "Unsupported architecture" >&2; exit 3 ;;
esac
powershell_package="powershell_${powershell_version}-1.deb_${powershell_arch}.deb"
curl -fL --retry 3 --silent --show-error \
  "https://github.com/PowerShell/PowerShell/releases/download/v${powershell_version}/${powershell_package}" \
  -o "/tmp/${powershell_package}"
apt-get install -y "/tmp/${powershell_package}"
rm -f "/tmp/${powershell_package}"

curl -fL --retry 3 --silent --show-error \
  "https://go.dev/dl/go${go_version}.linux-${go_arch}.tar.gz" -o /tmp/go.tar.gz
rm -rf /usr/local/go
tar -C /usr/local -xzf /tmp/go.tar.gz
rm -f /tmp/go.tar.gz
ln -sf /usr/local/go/bin/go /usr/local/bin/go
ln -sf /usr/local/go/bin/gofmt /usr/local/bin/gofmt
printf '%s\n' 'export PATH=/usr/local/go/bin:$PATH' > /etc/profile.d/go.sh

install -m 0755 -d /etc/docker
cat > /etc/docker/daemon.json <<'JSON'
{
  "storage-driver": "overlay2",
  "log-driver": "json-file",
  "log-opts": { "max-size": "10m", "max-file": "3" },
  "live-restore": true
}
JSON
usermod -aG docker "${linux_user}"
cat > /etc/wsl.conf <<EOF
[boot]
systemd=true

[user]
default=${linux_user}
EOF
systemctl enable docker.service containerd.service
rm -rf /var/lib/apt/lists/*
'@
    Invoke-WslBootstrap $DistroName $baseBootstrap @(
        $LinuxUser, $passwordWslPath, $PowerShellVersion, $GoVersion
    ) "Base WSL provisioning failed"
    Invoke-NativeChecked wsl.exe @("--terminate", $DistroName) "Could not terminate $DistroName"
    Start-Sleep -Seconds 2
    Invoke-NativeChecked wsl.exe @("--distribution", $DistroName, "--user", "root", "--exec", "systemctl", "start", "docker.service") "Docker did not start in $DistroName"

    if ($GitHubToken) { $plainToken = ConvertFrom-SecureValue $GitHubToken }
    elseif ($GitHubTokenFile) { $plainToken = [IO.File]::ReadAllText((Resolve-Path $GitHubTokenFile).Path).Trim() }
    else { $plainToken = Get-GhRunnerRegistrationToken -RepoUrl $RepositoryUrl }
    if ([string]::IsNullOrWhiteSpace($plainToken)) { throw "The runner registration token is empty." }
    $tokenTemp = New-RestrictedTemporaryFile $plainToken
    $plainToken = ""
    $tokenWslPath = ConvertTo-WslPath $tokenTemp $DistroName

    $runnerBootstrap = @'
set -Eeuo pipefail
linux_user="$1"
token_source_file="$2"
repo_url="$3"
runner_name="$4"
runner_labels="$5"
runner_version="$6"
runner_sha256="$7"
prune_until="$8"
prune_interval_seconds="$9"
[[ "${runner_sha256}" == "-" ]] && runner_sha256=""
runner_home="/home/${linux_user}/actions-runner"
work_root="${runner_home}/_work"
runtime_dir="/run/cabinet-runner"
token_file="/run/cabinet-runner-registration-token"

install -o "${linux_user}" -g "${linux_user}" -m 0400 /dev/null "${token_file}"
tr -d '\r\n' < "${token_source_file}" > "${token_file}"
trap 'rm -f "${token_file}"' EXIT
registration_token="$(cat "${token_file}")"
payload="$(jq -nc --arg url "${repo_url}" '{url: $url, runner_event: "registration"}')"
http_status="$(curl -sS -o /dev/null -w '%{http_code}' --connect-timeout 15 --max-time 30 \
  --request POST --header "Authorization: RemoteAuth ${registration_token}" \
  --header 'Content-Type: application/json' --data "${payload}" \
  https://api.github.com/actions/runner-registration || true)"
case "${http_status}" in
  200|201|204) ;;
  *) echo "GitHub runner registration token preflight failed (HTTP ${http_status})." >&2; exit 3 ;;
esac
unset registration_token

if [[ "${runner_version}" == "latest" ]]; then
  runner_version="$(curl -fsSL https://api.github.com/repos/actions/runner/releases/latest | jq -r .tag_name | sed 's/^v//')"
fi
case "$(uname -m)" in x86_64) runner_arch="x64" ;; aarch64|arm64) runner_arch="arm64" ;; *) exit 4 ;; esac
install -d -o "${linux_user}" -g "${linux_user}" -m 0750 "${runner_home}" "${work_root}"
archive="/tmp/actions-runner-${runner_version}.tar.gz"
curl -fL --retry 3 --silent --show-error \
  "https://github.com/actions/runner/releases/download/v${runner_version}/actions-runner-linux-${runner_arch}-${runner_version}.tar.gz" \
  -o "${archive}"
if [[ -n "${runner_sha256}" ]]; then printf '%s  %s\n' "${runner_sha256}" "${archive}" | sha256sum -c -; fi
tar -xzf "${archive}" -C "${runner_home}"
rm -f "${archive}"
"${runner_home}/bin/installdependencies.sh"
chown -R "${linux_user}:${linux_user}" "${runner_home}"

install -d -m 0755 /usr/local/libexec
cat > /usr/local/libexec/cabinet-runner-job-started.sh <<'HOOK'
#!/usr/bin/env bash
set -Eeuo pipefail
runtime_dir="${RUNNER_RUNTIME_DIR:-/run/cabinet-runner}"
exec 9>"${runtime_dir}/maintenance.lock"
flock -w 300 9
touch "${runtime_dir}/job-active"
HOOK
cat > /usr/local/libexec/cabinet-runner-job-completed.sh <<'HOOK'
#!/usr/bin/env bash
set -Eeuo pipefail
runtime_dir="${RUNNER_RUNTIME_DIR:-/run/cabinet-runner}"
exec 9>"${runtime_dir}/maintenance.lock"
flock -w 300 9
trap 'rm -f "${runtime_dir}/job-active"' EXIT
if [[ -n "${GITHUB_WORKSPACE:-}" && "${GITHUB_WORKSPACE}" == "${RUNNER_WORK_ROOT}/"* ]]; then
  timeout 5m rm -rf -- "${GITHUB_WORKSPACE}" || true
fi
timeout 10m docker system prune --all --force --filter "until=${RUNNER_PRUNE_UNTIL:-1h}" || true
timeout 5m docker volume prune --all --force || true
HOOK
cat > /usr/local/libexec/cabinet-runner-reaper <<'REAPER'
#!/usr/bin/env bash
set -Eeuo pipefail
runtime_dir="/run/cabinet-runner"
state_dir="/var/lib/cabinet-runner-reaper"
install -d -o "${RUNNER_USER}" -g "${RUNNER_USER}" -m 0750 "${runtime_dir}"
install -d -m 0755 "${state_dir}"
touch "${runtime_dir}/maintenance.lock"
chown "${RUNNER_USER}:${RUNNER_USER}" "${runtime_dir}/maintenance.lock"
exec 9>"${runtime_dir}/maintenance.lock"
flock -n 9 || exit 0
if [[ -e "${runtime_dir}/job-active" ]] || pgrep -u "${RUNNER_USER}" -f '/Runner.Worker' >/dev/null 2>&1; then
  echo "Runner job active; cleanup skipped."
  exit 0
fi
timeout 15m docker system prune --all --force --filter "until=${RUNNER_PRUNE_UNTIL}" || true
timeout 10m docker builder prune --all --force --filter "until=${RUNNER_PRUNE_UNTIL}" || true
timeout 5m docker volume prune --all --force || true
if [[ "${RUNNER_WORK_ROOT}" == /home/*/actions-runner/_work && -d "${RUNNER_WORK_ROOT}" ]]; then
  find "${RUNNER_WORK_ROOT}" -mindepth 1 -maxdepth 1 -exec rm -rf -- {} +
fi
journalctl --vacuum-size=100M >/dev/null 2>&1 || true
apt-get clean || true
stamp="${state_dir}/last-fstrim"
if [[ ! -e "${stamp}" ]] || find "${stamp}" -mmin +1440 -print -quit | grep -q .; then
  fstrim / || true
  touch "${stamp}"
fi
REAPER
chmod 0755 /usr/local/libexec/cabinet-runner-*

sudo -u "${linux_user}" bash -Eeuo pipefail -c '
  cd "$1"
  token_value="$(tr -d "\r\n" < "$3")"
  ./config.sh --unattended --url "$2" --token "${token_value}" --name "$4" --labels "$5" --work _work --replace
  unset token_value
' _ "${runner_home}" "${repo_url}" "${token_file}" "${runner_name}" "${runner_labels}"
[[ -x "${runner_home}/svc.sh" ]] || { echo "Runner configuration did not create svc.sh" >&2; exit 6; }
cd "${runner_home}"
./svc.sh install "${linux_user}"
runner_service="$(systemctl list-unit-files --type=service --no-legend 'actions.runner.*.service' | awk 'NR == 1 {print $1}')"
[[ -n "${runner_service}" ]] || { echo "Could not locate the installed runner service" >&2; exit 6; }
install -d -m 0755 "/etc/systemd/system/${runner_service}.d"
cat > "/etc/systemd/system/${runner_service}.d/10-cabinet.conf" <<EOF
[Service]
RuntimeDirectory=cabinet-runner
RuntimeDirectoryMode=0750
RuntimeDirectoryPreserve=yes
ExecStartPre=+/usr/bin/install -d -o ${linux_user} -g ${linux_user} -m 0750 ${runtime_dir}
ExecStartPre=+/usr/bin/touch ${runtime_dir}/maintenance.lock
ExecStartPre=+/usr/bin/chown ${linux_user}:${linux_user} ${runtime_dir}/maintenance.lock
Environment="ACTIONS_RUNNER_HOOK_JOB_STARTED=/usr/local/libexec/cabinet-runner-job-started.sh"
Environment="ACTIONS_RUNNER_HOOK_JOB_COMPLETED=/usr/local/libexec/cabinet-runner-job-completed.sh"
Environment="RUNNER_RUNTIME_DIR=${runtime_dir}"
Environment="RUNNER_WORK_ROOT=${work_root}"
Environment="RUNNER_PRUNE_UNTIL=${prune_until}"
EOF
cat > /etc/default/cabinet-runner-reaper <<EOF
RUNNER_USER=${linux_user}
RUNNER_WORK_ROOT=${work_root}
RUNNER_PRUNE_UNTIL=${prune_until}
EOF
cat > /etc/systemd/system/cabinet-runner-reaper.service <<'UNIT'
[Unit]
Description=Reap unused Cabinet runner Docker and workspace data
After=docker.service
Requires=docker.service
[Service]
Type=oneshot
EnvironmentFile=/etc/default/cabinet-runner-reaper
ExecStart=/usr/local/libexec/cabinet-runner-reaper
Nice=10
IOSchedulingClass=idle
TimeoutStartSec=30min
UNIT
cat > /etc/systemd/system/cabinet-runner-reaper.timer <<EOF
[Unit]
Description=Run idle-aware Cabinet runner cleanup
[Timer]
OnBootSec=10min
OnUnitActiveSec=${prune_interval_seconds}s
RandomizedDelaySec=2min
Persistent=true
Unit=cabinet-runner-reaper.service
[Install]
WantedBy=timers.target
EOF
systemctl daemon-reload
systemctl enable --now docker.service cabinet-runner-reaper.timer "${runner_service}"
'@
    $shaArgument = if ($RunnerSha256) { $RunnerSha256 } else { "-" }
    Invoke-WslBootstrap $DistroName $runnerBootstrap @(
        $LinuxUser, $tokenWslPath, $RepositoryUrl.TrimEnd('/'), $RunnerName,
        $RunnerLabels, $RunnerVersion, $shaArgument, $PruneUntil,
        [string]$PruneIntervalSeconds
    ) "GitHub runner provisioning failed"
} finally {
    $plainPassword = ""
    $plainToken = ""
    if ($passwordTemp) { Remove-Item -LiteralPath $passwordTemp -Force -ErrorAction SilentlyContinue }
    if ($tokenTemp) { Remove-Item -LiteralPath $tokenTemp -Force -ErrorAction SilentlyContinue }
}

if (-not $SkipAutostart) {
    Register-WslAutostartTask -Distribution $DistroName
}
Write-Host "Dedicated Cabinet WSL runner installed." -ForegroundColor Green
Write-Host ("Distribution: {0}`nRunner: {1}`nLocation: {2}" -f $DistroName, $RunnerName, $resolvedInstallLocation)
Write-Host ("Verify: wsl -d {0} -- systemctl status 'actions.runner.*'" -f $DistroName)
Write-Host ("Cleanup: wsl -d {0} -- systemctl list-timers cabinet-runner-reaper.timer" -f $DistroName)
