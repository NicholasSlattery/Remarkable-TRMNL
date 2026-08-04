$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$go = Join-Path $root '_tools\go\bin\go.exe'
$out = Join-Path $root 'dist\integration'
New-Item -ItemType Directory -Path $out -Force | Out-Null
$oldGOOS=$env:GOOS; $oldGOARCH=$env:GOARCH; $oldCGO=$env:CGO_ENABLED
try {
    $env:GOOS='linux'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
    & $go build -buildvcs=false -trimpath -ldflags '-s -w -X main.version=integration' -o (Join-Path $out 'trmnl-backend') "$root\backend\cmd\trmnl-remarkable"
    if ($LASTEXITCODE -ne 0) { throw 'integration backend build failed' }
    & $go build -buildvcs=false -trimpath -ldflags '-s -w' -o (Join-Path $out 'trmnl-mock') "$root\backend\cmd\trmnl-mock"
    if ($LASTEXITCODE -ne 0) { throw 'integration mock build failed' }
} finally {
    $env:GOOS=$oldGOOS; $env:GOARCH=$oldGOARCH; $env:CGO_ENABLED=$oldCGO
}
$linuxRoot = (wsl.exe wslpath -a $root.Replace('\','/')).Trim()
wsl.exe --exec python3 "$linuxRoot/tests/appload_harness.py" "$linuxRoot/dist/integration/trmnl-backend" "$linuxRoot/dist/integration/trmnl-mock"
if ($LASTEXITCODE -ne 0) { throw 'Linux AppLoad protocol integration failed' }
