#Requires -RunAsAdministrator
<#
.SYNOPSIS
    Installs Barry Allen speed test service from the latest GitHub release.
.DESCRIPTION
    Downloads the latest barryallen binary from GitHub releases,
    places it in C:\speedtest, and installs it as a Windows service.
.PARAMETER Arch
    Target architecture: amd64 (default) or arm64
.PARAMETER Uninstall
    Remove the service and optionally clean up files
#>

param(
    [ValidateSet("amd64", "arm64")]
    [string]$Arch = "amd64",
    [switch]$Uninstall
)

$ErrorActionPreference = "Stop"

$InstallDir = "C:\speedtest"
$BinaryName = "barryallen.exe"
$BinaryPath = Join-Path $InstallDir $BinaryName
$ServiceName = "BarryAllen"
$Repo = "martinsaul/barryallen"

function Get-LatestReleaseUrl {
    $apiUrl = "https://api.github.com/repos/$Repo/releases/latest"
    $release = Invoke-RestMethod -Uri $apiUrl -Headers @{ "User-Agent" = "BarryAllen-Installer" }
    $asset = $release.assets | Where-Object { $_.name -like "*windows-$Arch*" } | Select-Object -First 1
    if (-not $asset) {
        throw "No binary found for windows-$Arch in the latest release ($($release.tag_name))"
    }
    Write-Host "Found release $($release.tag_name) - $($asset.name)"
    return $asset.browser_download_url
}

function Install-BarryAllen {
    Write-Host "=== Barry Allen Speed Test Service Installer ===" -ForegroundColor Cyan
    Write-Host ""

    # Check if service already exists and stop it
    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        Write-Host "Existing service found. Stopping..."
        if ($existingService.Status -eq "Running") {
            Stop-Service -Name $ServiceName -Force
            Start-Sleep -Seconds 2
        }
        Write-Host "Removing existing service..."
        & $BinaryPath uninstall 2>$null
        Start-Sleep -Seconds 1
    }

    # Create install directory
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
        Write-Host "Created directory: $InstallDir"
    }

    # Download latest release
    Write-Host "Fetching latest release..."
    $downloadUrl = Get-LatestReleaseUrl
    Write-Host "Downloading from: $downloadUrl"
    Invoke-WebRequest -Uri $downloadUrl -OutFile $BinaryPath -UseBasicParsing
    Write-Host "Downloaded to: $BinaryPath" -ForegroundColor Green

    # Install and start service
    Write-Host "Installing service..."
    & $BinaryPath install
    if ($LASTEXITCODE -ne 0) { throw "Failed to install service" }

    Write-Host "Starting service..."
    & $BinaryPath start
    if ($LASTEXITCODE -ne 0) { throw "Failed to start service" }

    Write-Host ""
    Write-Host "=== Installation Complete ===" -ForegroundColor Green
    Write-Host "Binary:    $BinaryPath"
    Write-Host "CSV data:  $InstallDir\speedtest.csv"
    Write-Host "Log file:  $InstallDir\barryallen.log"
    Write-Host "Service:   $ServiceName (running, auto-start)"
    Write-Host ""
    Write-Host "Speed tests will run every 5 minutes." -ForegroundColor Cyan
    Write-Host "Run 'barryallen.exe run' from $InstallDir to test manually."
}

function Uninstall-BarryAllen {
    Write-Host "=== Barry Allen Uninstaller ===" -ForegroundColor Yellow

    $existingService = Get-Service -Name $ServiceName -ErrorAction SilentlyContinue
    if ($existingService) {
        if ($existingService.Status -eq "Running") {
            Write-Host "Stopping service..."
            Stop-Service -Name $ServiceName -Force
            Start-Sleep -Seconds 2
        }
        Write-Host "Removing service..."
        if (Test-Path $BinaryPath) {
            & $BinaryPath uninstall
        } else {
            sc.exe delete $ServiceName | Out-Null
        }
        Write-Host "Service removed." -ForegroundColor Green
    } else {
        Write-Host "Service not found, nothing to remove."
    }

    $removeFiles = Read-Host "Remove all files in $InstallDir including CSV data? (y/N)"
    if ($removeFiles -eq "y") {
        Remove-Item -Path $InstallDir -Recurse -Force
        Write-Host "Files removed." -ForegroundColor Green
    } else {
        # Just remove the binary
        if (Test-Path $BinaryPath) {
            Remove-Item -Path $BinaryPath -Force
            Write-Host "Binary removed. Data files kept in $InstallDir"
        }
    }

    Write-Host "Uninstall complete." -ForegroundColor Green
}

# Main
if ($Uninstall) {
    Uninstall-BarryAllen
} else {
    Install-BarryAllen
}
