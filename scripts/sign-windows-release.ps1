param(
    [Parameter(Mandatory=$true)][string]$Version,
    [Parameter(Mandatory=$true)][string]$PfxPath,
    [Parameter(Mandatory=$true)][string]$PfxPassword,
    [string]$TimestampUrl = 'http://timestamp.digicert.com',
    [switch]$AllowUntrusted,
    [string]$PublicCertificatePath = ''
)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
$packageRoot = Join-Path $root "build\release\TRMNL-for-reMarkable-$Version"
$installer = Join-Path $packageRoot 'TRMNL Installer.exe'
$releaseDir = Join-Path $root 'release'
$zip = Join-Path $releaseDir "TRMNL-for-reMarkable-$Version-Windows-x64.zip"

if (-not (Test-Path -LiteralPath $PfxPath)) { throw "Signing certificate not found: $PfxPath" }
if (-not (Test-Path -LiteralPath $installer)) { throw "Installer not found: $installer" }

$signToolCommand = Get-Command signtool.exe -ErrorAction SilentlyContinue
if ($signToolCommand) {
    $signTool = $signToolCommand.Source
} else {
    $kits = Join-Path ${env:ProgramFiles(x86)} 'Windows Kits\10\bin'
    $signTool = Get-ChildItem -Path (Join-Path $kits '*\x64\signtool.exe') -ErrorAction SilentlyContinue |
        Sort-Object FullName -Descending | Select-Object -First 1 -ExpandProperty FullName
}
if (-not $signTool) { throw 'signtool.exe was not found in PATH or the Windows 10 SDK.' }

& $signTool sign /fd SHA256 /td SHA256 /tr $TimestampUrl /f $PfxPath /p $PfxPassword $installer
if ($LASTEXITCODE -ne 0) { throw 'Authenticode signing failed.' }
$signature = Get-AuthenticodeSignature -LiteralPath $installer
if (-not $signature.SignerCertificate) { throw 'The signed installer has no Authenticode signer certificate.' }

if ($AllowUntrusted) {
    $expected = New-Object Security.Cryptography.X509Certificates.X509Certificate2($PfxPath, $PfxPassword)
    $untrustedRoot = $signature.Status -eq 'UnknownError' -and $signature.StatusMessage -match 'root certificate.*not trusted'
    if ($signature.SignatureType -ne 'Authenticode' -or
        $signature.SignerCertificate.Thumbprint -ne $expected.Thumbprint -or
        -not $signature.TimeStamperCertificate -or
        -not $untrustedRoot) {
        throw "Self-signed Authenticode verification failed: $($signature.Status) $($signature.StatusMessage)"
    }
} else {
    & $signTool verify /pa /all /v $installer
    if ($LASTEXITCODE -ne 0) { throw 'Authenticode signature verification failed.' }
}

if ($PublicCertificatePath) {
    if (-not (Test-Path -LiteralPath $PublicCertificatePath)) { throw "Public certificate not found: $PublicCertificatePath" }
    Copy-Item -LiteralPath $PublicCertificatePath -Destination (Join-Path $packageRoot 'SELF-SIGNED-CERTIFICATE.cer') -Force
    Copy-Item -LiteralPath (Join-Path $root 'docs\code-signing.md') -Destination (Join-Path $packageRoot 'SELF-SIGNED-SIGNATURE.md') -Force
    $certificateHash = (Get-FileHash -Algorithm SHA256 -LiteralPath $PublicCertificatePath).Hash.ToLowerInvariant()
    $certificateDetails = @(
        "Subject: $($signature.SignerCertificate.Subject)"
        "Thumbprint (SHA-1 identifier): $($signature.SignerCertificate.Thumbprint)"
        "Public certificate SHA-256: $certificateHash"
        "Valid from: $($signature.SignerCertificate.NotBefore.ToUniversalTime().ToString('o'))"
        "Valid until: $($signature.SignerCertificate.NotAfter.ToUniversalTime().ToString('o'))"
    ) -join [Environment]::NewLine
    Set-Content -Encoding ascii -LiteralPath (Join-Path $packageRoot 'SELF-SIGNED-CERTIFICATE.txt') -Value $certificateDetails
}

$goCommand = Get-Command go -ErrorAction SilentlyContinue
$go = if ($goCommand) { $goCommand.Source } else { Join-Path $root '_tools\go\bin\go.exe' }
if (Test-Path -LiteralPath $zip) { Remove-Item -LiteralPath $zip }
& $go run (Join-Path $PSScriptRoot 'package-zip.go') $packageRoot $zip
if ($LASTEXITCODE -ne 0) { throw 'Signed release ZIP creation failed.' }
$hash = (Get-FileHash -Algorithm SHA256 -LiteralPath $zip).Hash.ToLowerInvariant()
Set-Content -Encoding ascii -LiteralPath (Join-Path $releaseDir 'SHA256SUMS.txt') -Value "$hash  $(Split-Path -Leaf $zip)"
Write-Host "Authenticode-signed release ready: $zip"
Write-Host "SHA-256: $hash"
