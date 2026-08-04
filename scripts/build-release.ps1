param([string]$Version = '')
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
function Write-Utf8NoBom([string]$Path, [string]$Value) {
    [IO.File]::WriteAllText($Path, $Value, (New-Object Text.UTF8Encoding($false)))
}
if (-not $Version) { $Version = (Get-Content -Raw -LiteralPath (Join-Path $root 'VERSION')).Trim() }
if ($Version -notmatch '^\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$') { throw "Invalid version: $Version" }

& (Join-Path $PSScriptRoot 'build.ps1') -Version $Version
& (Join-Path $PSScriptRoot 'fetch-runtime.ps1')

$goCommand = Get-Command go -ErrorAction SilentlyContinue
$go = if ($goCommand) { $goCommand.Source } else { Join-Path $root '_tools\go\bin\go.exe' }
$work = Join-Path $root 'build\release'
$packageRoot = Join-Path $work "TRMNL-for-reMarkable-$Version"
$payload = Join-Path $packageRoot 'payload'
if (Test-Path -LiteralPath $work) { Remove-Item -Recurse -LiteralPath $work }
New-Item -ItemType Directory -Path (Join-Path $payload 'appload'),(Join-Path $payload 'licenses') -Force | Out-Null

$deviceArchive = Join-Path $payload 'trmnl-remarkable-device.tar.gz'
& $go run (Join-Path $PSScriptRoot 'package-device.go') $root $deviceArchive
if ($LASTEXITCODE -ne 0) { throw 'Device package creation failed' }
$deviceHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $deviceArchive).Hash.ToLowerInvariant()
Set-Content -NoNewline -Encoding ascii -LiteralPath (Join-Path $payload 'trmnl-remarkable-device.sha256') -Value $deviceHash

$downloads = Join-Path $root '_downloads'
Copy-Item -LiteralPath (Join-Path $downloads 'xovi-aarch64.tar.gz') -Destination $payload
Copy-Item -LiteralPath (Join-Path $downloads 'appload-release\appload.so') -Destination (Join-Path $payload 'appload')
Copy-Item -LiteralPath (Join-Path $downloads 'appload-release\shims\qtfb-shim.so') -Destination (Join-Path $payload 'appload')
Copy-Item -LiteralPath (Join-Path $downloads 'appload-release\shims\qtfb-shim-32bit.so') -Destination (Join-Path $payload 'appload')
Copy-Item -LiteralPath (Join-Path $downloads 'licenses\XOVI-LICENSE') -Destination (Join-Path $payload 'licenses')
Copy-Item -LiteralPath (Join-Path $downloads 'licenses\APPLOAD-LICENSE') -Destination (Join-Path $payload 'licenses')
Copy-Item -LiteralPath (Join-Path $downloads 'licenses\EXTENSIONS-LICENSE') -Destination (Join-Path $payload 'licenses')
Copy-Item -LiteralPath (Join-Path $root 'scripts\install-device-runtime.sh') -Destination $payload

$files = [ordered]@{}
Get-ChildItem -Recurse -File -LiteralPath $payload | Where-Object Name -ne 'manifest.json' | Sort-Object FullName | ForEach-Object {
    $relative = $_.FullName.Substring($payload.Length + 1).Replace('\','/')
    $files[$relative] = (Get-FileHash -Algorithm SHA256 -LiteralPath $_.FullName).Hash.ToLowerInvariant()
}
$manifestJson = [ordered]@{ version=$Version; files=$files } | ConvertTo-Json -Depth 4
Write-Utf8NoBom (Join-Path $payload 'manifest.json') $manifestJson

$oldGOOS=$env:GOOS; $oldGOARCH=$env:GOARCH; $oldCGO=$env:CGO_ENABLED
try {
    $env:GOOS='windows'; $env:GOARCH='amd64'; $env:CGO_ENABLED='0'
    & $go build -buildvcs=false -trimpath -ldflags "-s -w -H=windowsgui -X main.version=$Version" -o (Join-Path $packageRoot 'TRMNL Installer.exe') (Join-Path $root 'installer')
    if ($LASTEXITCODE -ne 0) { throw 'Graphical installer build failed' }
} finally {
    $env:GOOS=$oldGOOS; $env:GOARCH=$oldGOARCH; $env:CGO_ENABLED=$oldCGO
}

Copy-Item -LiteralPath (Join-Path $root 'EASY_INSTALL.md') -Destination (Join-Path $packageRoot 'READ ME - Install TRMNL.md')
Copy-Item -LiteralPath (Join-Path $root 'LICENSE'),(Join-Path $root 'THIRD_PARTY_NOTICES.md') -Destination $packageRoot
$sourceOut = Join-Path $packageRoot 'Corresponding Source'
New-Item -ItemType Directory -Path $sourceOut -Force | Out-Null
Copy-Item -Path (Join-Path $downloads 'sources\*.tar.gz') -Destination $sourceOut

$goComponents = @()
& $go list -m -f '{{if not .Main}}{{.Path}}|{{.Version}}{{end}}' all | Where-Object { $_ } | ForEach-Object {
    $parts = $_ -split '\|', 2
    $goComponents += [ordered]@{type='library';name=$parts[0];version=$parts[1];licenses=@(@{license=@{id='BSD-3-Clause'}})}
}
$components = @($goComponents) + @(
    [ordered]@{type='framework';name='XOVI';version='0.3.3';licenses=@(@{license=@{id='LGPL-3.0-only'}})},
    [ordered]@{type='application';name='AppLoad';version='0.5.3';licenses=@(@{license=@{id='GPL-3.0-only'}})},
    [ordered]@{type='framework';name='rm-xovi-extensions';version='19';licenses=@(@{license=@{id='GPL-3.0-only'}})}
)
$sbom = [ordered]@{
    bomFormat='CycloneDX'; specVersion='1.5'; version=1
    metadata=[ordered]@{ component=[ordered]@{ type='application'; name='trmnl-remarkable'; version=$Version } }
    components=$components
}
Write-Utf8NoBom (Join-Path $packageRoot 'SBOM.cdx.json') ($sbom | ConvertTo-Json -Depth 8)

$releaseDir = Join-Path $root 'release'
New-Item -ItemType Directory -Path $releaseDir -Force | Out-Null
$zip = Join-Path $releaseDir "TRMNL-for-reMarkable-$Version-Windows-x64.zip"
if (Test-Path -LiteralPath $zip) { Remove-Item -LiteralPath $zip }
& $go run (Join-Path $PSScriptRoot 'package-zip.go') $packageRoot $zip
if ($LASTEXITCODE -ne 0) { throw 'Release ZIP creation failed' }
$hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $zip).Hash.ToLowerInvariant()
Set-Content -Encoding ascii -LiteralPath (Join-Path $releaseDir 'SHA256SUMS.txt') -Value "$hash  $(Split-Path -Leaf $zip)"
Write-Host "Release ready: $zip"
Write-Host "SHA-256: $hash"
