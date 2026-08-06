param([string]$Destination = (Join-Path (Split-Path -Parent $PSScriptRoot) '_downloads'))
$ErrorActionPreference = 'Stop'

$assets = @(
    @{
        Name = 'xovi-aarch64.tar.gz'
        Url = 'https://github.com/asivery/rm-xovi-extensions/releases/download/v19-23052026/xovi-aarch64.tar.gz'
        SHA256 = '32d64d1262ddc984e3235c7d0340a398fe6d5b3efa6a979865f5977b32630d27'
    },
    @{
        Name = 'appload-aarch64.zip'
        Url = 'https://github.com/asivery/rm-appload/releases/download/v0.5.3/appload-aarch64.zip'
        SHA256 = '032e3f2c57a004aba4425894758e4b542c67590efd222e3b3d5141124c45e84d'
    },
    @{
        Name = 'sources/xovi-v0.3.3-source.tar.gz'
        Url = 'https://github.com/asivery/xovi/archive/refs/tags/v0.3.3.tar.gz'
        SHA256 = '2ac3d4ee46851737b8dc6de699073dd612432122f057f1c75c8489678b3fb29f'
    },
    @{
        Name = 'sources/rm-appload-v0.5.3-source.tar.gz'
        Url = 'https://github.com/asivery/rm-appload/archive/refs/tags/v0.5.3.tar.gz'
        SHA256 = 'df5878cf06e1167b9156c2455bab762366d2a4b2058179073734d9ca72e46e94'
    },
    @{
        Name = 'sources/rm-xovi-extensions-v19-source.tar.gz'
        Url = 'https://github.com/asivery/rm-xovi-extensions/archive/refs/tags/v19-23052026.tar.gz'
        SHA256 = '06460210d74779cea287c33ae99408b286908b4bc6c8ae8e9e6b8ca39d19fd80'
    },
    # Redistributed verbatim with the release, so they are pinned like every
    # other fetched artefact rather than trusted at download time.
    @{
        Name = 'licenses/XOVI-LICENSE'
        Url = 'https://raw.githubusercontent.com/asivery/xovi/v0.3.3/LICENSE'
        SHA256 = '3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986'
    },
    @{
        Name = 'licenses/APPLOAD-LICENSE'
        Url = 'https://raw.githubusercontent.com/asivery/rm-appload/v0.5.3/LICENSE'
        SHA256 = '3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986'
    },
    @{
        Name = 'licenses/EXTENSIONS-LICENSE'
        Url = 'https://raw.githubusercontent.com/asivery/rm-xovi-extensions/v19-23052026/LICENSE'
        SHA256 = '3972dc9744f6499f0f9b2dbf76696f2ae7ad8af9b23dde66d6af86c9dfb36986'
    }
)

New-Item -ItemType Directory -Path $Destination -Force | Out-Null
foreach ($asset in $assets) {
    $path = Join-Path $Destination $asset.Name
    New-Item -ItemType Directory -Path (Split-Path -Parent $path) -Force | Out-Null
    if (-not (Test-Path -LiteralPath $path) -or (Get-FileHash -Algorithm SHA256 -LiteralPath $path).Hash -ne $asset.SHA256) {
        Write-Host "Downloading $($asset.Name)..."
        Invoke-WebRequest -UseBasicParsing -Uri $asset.Url -OutFile $path
    }
    $actual = (Get-FileHash -Algorithm SHA256 -LiteralPath $path).Hash.ToLowerInvariant()
    if ($actual -ne $asset.SHA256) { throw "Checksum mismatch for $($asset.Name)" }
}

$appLoadOut = Join-Path $Destination 'appload-release'
if (Test-Path -LiteralPath $appLoadOut) { Remove-Item -Recurse -LiteralPath $appLoadOut }
Expand-Archive -LiteralPath (Join-Path $Destination 'appload-aarch64.zip') -DestinationPath $appLoadOut

Write-Host "Verified runtime assets are ready in $Destination"
