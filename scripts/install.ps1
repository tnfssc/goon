param(
    [string]$InstallDir = "$env:LOCALAPPDATA\Programs\goon"
)

$ErrorActionPreference = "Stop"

$Repo = "tnfssc/goon"
$DownloadBase = "https://github.com/$Repo/releases/latest/download"
$BinaryName = "goon.exe"

function Get-Architecture {
    switch ([System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture) {
        "Arm64" { return "arm64" }
        "X64" { return "amd64" }
        default { throw "Unsupported architecture $_" }
    }
}

$arch = Get-Architecture
$assetName = "goon-windows-$arch.exe"
$url = "$DownloadBase/$assetName"

$tempFile = Join-Path $env:TEMP ("$assetName-" + [System.Guid]::NewGuid().ToString())

Write-Host "Downloading $url"
Invoke-WebRequest -Uri $url -OutFile $tempFile -UseBasicParsing

if (-not (Test-Path $InstallDir)) {
    Write-Host "Creating $InstallDir"
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$target = Join-Path $InstallDir $BinaryName
Move-Item -Path $tempFile -Destination $target -Force

Write-Host "goon installed to $target"

$pathEntries = $env:PATH -split ';'
if ($pathEntries -notcontains $InstallDir) {
    Write-Warning "$InstallDir is not on your PATH. Add it via System Properties or run:`n  setx PATH \"$InstallDir;%PATH%\""
}

Write-Host "Run 'goon --help' from a new PowerShell window to verify."