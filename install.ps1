# Gito / gt Installer for Windows (PowerShell)
# Usage: irm https://raw.githubusercontent.com/mijanlab/gito/main/install.ps1 | iex

$ErrorActionPreference = "Stop"

$Repo = "mijanlab/gito"
$BinaryName = "gito.exe"
$ShortName = "gt.exe"
$InstallDir = "$env:LOCALAPPDATA\Programs\gito"

Write-Host "◈ Installing gt / Gito — Premium Interactive Git TUI..." -ForegroundColor Cyan

# 1. Detect Architecture
$Arch = "amd64"
if ($env:PROCESSOR_ARCHITECTURE -eq "ARM64") {
    $Arch = "arm64"
}

# 2. Ensure install directory exists
if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Path $InstallDir -Force | Out-Null
}

$TargetPath = Join-Path $InstallDir $BinaryName
$ShortPath = Join-Path $InstallDir $ShortName

# 3. Check if local Go build is possible
$GoInstalled = (Get-Command go -ErrorAction SilentlyContinue)
if (Test-Path "go.mod" -and (Test-Path "cmd\gito") -and $GoInstalled) {
    Write-Host "Building from source with local Go..." -ForegroundColor Gray
    go build -ldflags="-s -w" -o $TargetPath .\cmd\gito
    go build -ldflags="-s -w" -o $ShortPath .\cmd\gt
    Write-Host "✓ Gito & gt built and installed to $InstallDir" -ForegroundColor Green
} elseif ($GoInstalled) {
    Write-Host "Installing via 'go install'..." -ForegroundColor Gray
    $env:GOBIN = $InstallDir
    go install github.com/mijanlab/gito/cmd/gito@latest
    go install github.com/mijanlab/gito/cmd/gt@latest
    Write-Host "✓ Gito & gt installed to $InstallDir" -ForegroundColor Green
} else {
    # Download binary from GitHub Releases
    $LatestTag = "v0.0.5"
    try {
        $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
        if ($ReleaseInfo.tag_name) {
            $LatestTag = $ReleaseInfo.tag_name
        }
    } catch {
        # Fallback to default tag
    }

    $DownloadUrl = "https://github.com/$Repo/releases/download/$LatestTag/gito-windows-$Arch.exe"
    Write-Host "Downloading precompiled binary from $DownloadUrl..." -ForegroundColor Gray

    try {
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $TargetPath -UseBasicParsing
        Copy-Item -Path $TargetPath -Destination $ShortPath -Force
        Write-Host "✓ Gito & gt downloaded and installed to $InstallDir" -ForegroundColor Green
    } catch {
        Write-Error "Failed to download binary. Please install Go and run 'go install github.com/$Repo/cmd/gt@latest'."
        exit 1
    }
}

# 4. Check / Update PATH
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    Write-Host "Adding $InstallDir to user PATH..." -ForegroundColor Gray
    $NewPath = "$UserPath;$InstallDir"
    [Environment]::SetEnvironmentVariable("PATH", $NewPath, "User")
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "✓ User PATH updated successfully." -ForegroundColor Green
}

Write-Host "Done! Restart your terminal or PowerShell, then type 'gt' to run." -ForegroundColor Cyan
