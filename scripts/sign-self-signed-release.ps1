param(
    [string]$Version = '',
    [string]$Subject = 'CN=TRMNL for reMarkable Community Release'
)
$ErrorActionPreference = 'Stop'
$root = Split-Path -Parent $PSScriptRoot
if (-not $Version) { $Version = (Get-Content -Raw -LiteralPath (Join-Path $root 'VERSION')).Trim() }
if ($Version -notmatch '^\d+\.\d+\.\d+([-.][0-9A-Za-z.-]+)?$') { throw "Invalid version: $Version" }

$packageRoot = Join-Path $root "build\release\TRMNL-for-reMarkable-$Version"
if (-not (Test-Path -LiteralPath (Join-Path $packageRoot 'TRMNL Installer.exe'))) {
    throw "Build the release before signing: scripts/build-release.ps1 -Version $Version"
}

$cert = Get-ChildItem Cert:\CurrentUser\My -CodeSigningCert | Where-Object {
    $_.Subject -eq $Subject -and $_.HasPrivateKey -and $_.NotAfter -gt (Get-Date).AddDays(30)
} | Sort-Object NotAfter -Descending | Select-Object -First 1
if (-not $cert) {
    $cert = New-SelfSignedCertificate -Type CodeSigningCert -Subject $Subject `
        -FriendlyName 'TRMNL for reMarkable self-signed release certificate' `
        -CertStoreLocation Cert:\CurrentUser\My -KeyAlgorithm RSA -KeyLength 3072 `
        -HashAlgorithm SHA256 -KeyExportPolicy Exportable -NotAfter (Get-Date).AddYears(3)
}

$temp = Join-Path ([IO.Path]::GetTempPath()) ("trmnl-sign-" + [guid]::NewGuid().ToString('N'))
New-Item -ItemType Directory -Path $temp | Out-Null
$pfx = Join-Path $temp 'signing.pfx'
$cer = Join-Path $temp 'signing.cer'
$random = New-Object byte[] 32
[Security.Cryptography.RandomNumberGenerator]::Create().GetBytes($random)
$plainPassword = [Convert]::ToBase64String($random)
$securePassword = ConvertTo-SecureString $plainPassword -AsPlainText -Force
try {
    Export-PfxCertificate -Cert $cert -FilePath $pfx -Password $securePassword | Out-Null
    Export-Certificate -Cert $cert -FilePath $cer -Type CERT | Out-Null
    & (Join-Path $PSScriptRoot 'sign-windows-release.ps1') -Version $Version `
        -PfxPath $pfx -PfxPassword $plainPassword -AllowUntrusted -PublicCertificatePath $cer
    if ($LASTEXITCODE -ne 0) { throw 'Self-signed release packaging failed.' }
} finally {
    $plainPassword = $null
    $securePassword = $null
    if (Test-Path -LiteralPath $temp) { Remove-Item -Recurse -Force -LiteralPath $temp }
}

Write-Host "Self-signed certificate: $($cert.Subject)"
Write-Host "Certificate thumbprint: $($cert.Thumbprint)"
Write-Warning 'This certificate is not publicly trusted; Windows warnings remain until a user explicitly trusts it.'
