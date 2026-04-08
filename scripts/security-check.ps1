$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$env:GOCACHE = Join-Path $root ".gocache"

New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null

Write-Host "[1/4] go build"
go build ./cmd/pdfsplitter
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[2/4] go test"
go test ./...
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

Write-Host "[3/4] PDFBox 확인"
$pdfbox = Join-Path $root "tools\pdfbox\pdfbox-app-3.0.7.jar"
if (Test-Path $pdfbox) {
  Get-Item $pdfbox | Select-Object FullName,Length,LastWriteTime
} else {
  Write-Warning "최신 PDFBox 3.0.7 JAR가 없습니다. 썸네일 보안 하드닝이 구버전 fallback으로 동작할 수 있습니다."
}

Write-Host "[4/4] govulncheck"
$govulncheck = Get-Command govulncheck -ErrorAction SilentlyContinue
if ($govulncheck) {
  & $govulncheck.Source ./...
  if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }
} else {
  Write-Warning "govulncheck가 설치되어 있지 않아 취약점 스캔은 건너뜁니다."
}

Write-Host "Security check completed."
