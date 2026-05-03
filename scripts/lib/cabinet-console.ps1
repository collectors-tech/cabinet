$script:CabinetConsoleColorEnabled = -not [bool]$env:NO_COLOR

function Write-CabinetHost {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Message,
    [ConsoleColor]$Color = [ConsoleColor]::Gray
  )

  if ($script:CabinetConsoleColorEnabled) {
    Write-Host $Message -ForegroundColor $Color
    return
  }

  Write-Host $Message
}

function Write-CabinetBanner {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Command,
    [string]$Summary = ''
  )

  Write-Host ''
  Write-CabinetHost 'CABINET' -Color Cyan
  Write-CabinetHost ("Command: {0}" -f $Command) -Color White
  if ($Summary.Trim() -ne '') {
    Write-CabinetHost ("Purpose: {0}" -f $Summary.Trim()) -Color DarkGray
  }
  if ($env:CI) {
    Write-CabinetHost 'Mode: CI log compatible' -Color DarkGray
  }
}

function Write-CabinetSection {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Title
  )

  Write-Host ''
  Write-CabinetHost ("== {0} ==" -f $Title) -Color Cyan
}

function Write-CabinetStatus {
  param(
    [ValidateSet('info', 'run', 'ok', 'warn', 'error')]
    [string]$State = 'info',
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  $label = $State.ToUpperInvariant().PadRight(5)
  $color = switch ($State) {
    'ok' { [ConsoleColor]::Green }
    'warn' { [ConsoleColor]::Yellow }
    'error' { [ConsoleColor]::Red }
    'run' { [ConsoleColor]::Cyan }
    default { [ConsoleColor]::Gray }
  }

  Write-CabinetHost ("[{0}] {1}" -f $label, $Message) -Color $color
}

function Write-CabinetKeyValue {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Key,
    [AllowEmptyString()]
    [string]$Value = ''
  )

  Write-CabinetHost ("  {0,-12} {1}" -f ($Key + ':'), $Value) -Color Gray
}

function Write-CabinetHint {
  param(
    [Parameter(Mandatory = $true)]
    [string]$Message
  )

  Write-CabinetHost ("  hint: {0}" -f $Message) -Color DarkGray
}
