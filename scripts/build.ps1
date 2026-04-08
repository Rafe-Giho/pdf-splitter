$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$dist = Join-Path $root "dist\\windows"
$output = Join-Path $dist "pdf-splitter.exe"
$rootOutput = Join-Path $root "pdfsplitter.exe"
$pdfboxDist = Join-Path $dist "tools\\pdfbox"
$env:GOCACHE = Join-Path $root ".gocache"

New-Item -ItemType Directory -Force -Path $dist | Out-Null
New-Item -ItemType Directory -Force -Path $pdfboxDist | Out-Null
New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

$env:CGO_ENABLED = "0"

go run ./tools/icon/genicon.go
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

$icon = Join-Path $root "assets\\icon.ico"
$syso = Join-Path $root "cmd\\pdfsplitter\\rsrc_windows.syso"
$rsrc = Get-Command rsrc -ErrorAction SilentlyContinue

if (-not $rsrc) {
  $gopath = (go env GOPATH).Trim()
  $candidate = Join-Path $gopath "bin\\rsrc.exe"
  if (Test-Path $candidate) {
    $rsrc = Get-Item $candidate
  }
}

if ($rsrc) {
  & $rsrc.Source -ico $icon -o $syso
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} elseif (-not (Test-Path $syso)) {
  Write-Warning "rsrc 도구가 없어 아이콘 리소스를 새로 만들지 못했습니다."
}

go build -trimpath -ldflags "-s -w -H windowsgui" -o $output ./cmd/pdfsplitter
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Copy-Item -LiteralPath $output -Destination $rootOutput -Force

$pdfboxCandidates = @(
  (Join-Path $root "tools\\pdfbox\\pdfbox-app-3.0.7.jar"),
  (Join-Path $root "tools\\pdfbox\\pdfbox-app-3.0.3.jar")
)
foreach ($candidate in $pdfboxCandidates) {
  if (Test-Path $candidate) {
    Copy-Item -Path $candidate -Destination (Join-Path $pdfboxDist (Split-Path $candidate -Leaf)) -Force
    break
  }
}

Write-Host "Built: $output"
Write-Host "Synced: $rootOutput"
