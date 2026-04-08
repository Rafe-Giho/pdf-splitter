param(
  [string]$Version = ""
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $PSScriptRoot
$env:GOCACHE = Join-Path $root ".gocache"

if ([string]::IsNullOrWhiteSpace($Version)) {
  $Version = (Get-Content (Join-Path $root "VERSION") -Raw).Trim()
}
if (-not $Version) {
  throw "VERSION 값을 확인할 수 없습니다."
}

$buildScript = Join-Path $root "scripts\build.ps1"
$releaseRoot = Join-Path $root "dist\release\$Version"
$stageRoot = Join-Path $releaseRoot "pdf-splitter-$Version-windows-x64"
$archivePath = Join-Path $releaseRoot "pdf-splitter-$Version-windows-x64.zip"
$checksumPath = Join-Path $releaseRoot "SHA256SUMS.txt"
$releaseNotesSource = Join-Path $root ("releases\" + $Version + ".md")
$releaseNotesTarget = Join-Path $releaseRoot "RELEASE_NOTES.md"

New-Item -ItemType Directory -Force -Path $env:GOCACHE | Out-Null
New-Item -ItemType Directory -Force -Path $releaseRoot | Out-Null

& $buildScript
if ($LASTEXITCODE -ne 0) { exit $LASTEXITCODE }

if (Test-Path $stageRoot) {
  Remove-Item -LiteralPath $stageRoot -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $stageRoot | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $stageRoot "tools\pdfbox") | Out-Null
New-Item -ItemType Directory -Force -Path (Join-Path $stageRoot "assets") | Out-Null

$exePath = Join-Path $root "dist\windows\pdf-splitter.exe"
if (-not (Test-Path $exePath)) {
  throw "Windows 실행 파일을 찾을 수 없습니다: $exePath"
}
Copy-Item -LiteralPath $exePath -Destination (Join-Path $stageRoot "pdf-splitter.exe") -Force

$pdfboxCandidates = @(
  (Join-Path $root "dist\windows\tools\pdfbox\pdfbox-app-3.0.7.jar"),
  (Join-Path $root "dist\windows\tools\pdfbox\pdfbox-app-3.0.3.jar"),
  (Join-Path $root "tools\pdfbox\pdfbox-app-3.0.7.jar"),
  (Join-Path $root "tools\pdfbox\pdfbox-app-3.0.3.jar")
)
$pdfboxJar = $null
foreach ($candidate in $pdfboxCandidates) {
  if (Test-Path $candidate) {
    $pdfboxJar = $candidate
    break
  }
}
if (-not $pdfboxJar) {
  throw "PDFBox JAR를 찾을 수 없습니다."
}
$pdfboxStageDir = Join-Path $stageRoot "tools\pdfbox"
$pdfboxStagePath = Join-Path $pdfboxStageDir (Split-Path $pdfboxJar -Leaf)
Copy-Item -LiteralPath $pdfboxJar -Destination $pdfboxStagePath -Force

Copy-Item -LiteralPath (Join-Path $root "README.md") -Destination (Join-Path $stageRoot "README.md") -Force
Copy-Item -LiteralPath (Join-Path $root "CHANGELOG.md") -Destination (Join-Path $stageRoot "CHANGELOG.md") -Force
Copy-Item -LiteralPath (Join-Path $root "VERSION") -Destination (Join-Path $stageRoot "VERSION") -Force
Copy-Item -LiteralPath (Join-Path $root "assets\icon.png") -Destination (Join-Path $stageRoot "assets\icon.png") -Force
Copy-Item -LiteralPath (Join-Path $root "assets\readme-preview.svg") -Destination (Join-Path $stageRoot "assets\readme-preview.svg") -Force
Copy-Item -LiteralPath (Join-Path $root "assets\readme-preview.png") -Destination (Join-Path $stageRoot "assets\readme-preview.png") -Force
Copy-Item -LiteralPath (Join-Path $root "LICENSE") -Destination (Join-Path $stageRoot "LICENSE") -Force
Copy-Item -LiteralPath (Join-Path $root "NOTICE") -Destination (Join-Path $stageRoot "NOTICE") -Force
Copy-Item -LiteralPath (Join-Path $root "THIRD_PARTY_NOTICES.md") -Destination (Join-Path $stageRoot "THIRD_PARTY_NOTICES.md") -Force

if (Test-Path $releaseNotesSource) {
  Copy-Item -LiteralPath $releaseNotesSource -Destination $releaseNotesTarget -Force
}

if (Test-Path $archivePath) {
  Remove-Item -LiteralPath $archivePath -Force
}
Compress-Archive -Path (Join-Path $stageRoot "*") -DestinationPath $archivePath -CompressionLevel Optimal

$zipHash = (Get-FileHash -Algorithm SHA256 $archivePath).Hash.ToLower()
$exeHash = (Get-FileHash -Algorithm SHA256 (Join-Path $stageRoot "pdf-splitter.exe")).Hash.ToLower()
$jarHash = (Get-FileHash -Algorithm SHA256 $pdfboxStagePath).Hash.ToLower()

@(
  "$zipHash *$(Split-Path $archivePath -Leaf)"
  "$exeHash *pdf-splitter.exe"
  "$jarHash *tools/pdfbox/$(Split-Path $pdfboxJar -Leaf)"
) | Set-Content -Path $checksumPath -Encoding ascii

Write-Host "Release package created:"
Write-Host "  Archive   : $archivePath"
Write-Host "  Checksums : $checksumPath"
if (Test-Path $releaseNotesTarget) {
  Write-Host "  Notes     : $releaseNotesTarget"
}
