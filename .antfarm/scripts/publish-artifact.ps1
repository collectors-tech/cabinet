param(
  [string]$Name = "cabinet",
  [string]$ValidationSummary = "",
  [string]$ArtifactPath = ""
)

$ErrorActionPreference = "Stop"
$repoRoot = Resolve-Path (Join-Path $PSScriptRoot "..\..")
Push-Location $repoRoot
try {
  $sha = (git rev-parse --short HEAD).Trim()
  $ts = Get-Date -Format "yyyyMMdd-HHmmss"
  $artifactId = "release-$Name-$sha-$ts"

  $dir = Join-Path $repoRoot ".antfarm\artifacts"
  New-Item -ItemType Directory -Path $dir -Force | Out-Null

  $manifest = [ordered]@{
    artifactId = $artifactId
    repo = "collectors-tech/cabinet"
    gitSha = $sha
    createdAt = (Get-Date).ToString("o")
    artifactPath = $ArtifactPath
    validationSummary = $ValidationSummary
  } | ConvertTo-Json -Depth 6

  $out = Join-Path $dir "$artifactId.json"
  Set-Content -Path $out -Value $manifest -Encoding UTF8
  Write-Output "Published artifact manifest: $out"
}
finally { Pop-Location }
