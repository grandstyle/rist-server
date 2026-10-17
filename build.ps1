$ErrorActionPreference = "Stop"

$env:PATH = "C:\Program Files\Go\bin;$env:USERPROFILE\go\bin;$env:PATH"

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "  Building RIST Receiver & Setup (Go)   " -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

# 0. Ensure assets\ffmpeg.exe is present (auto-download if building from clean git clone)
if (-not (Test-Path "assets\ffmpeg.exe")) {
    Write-Host "[0/4] Downloading embedded FFmpeg engine (librist) from GitHub Releases..." -ForegroundColor Yellow
    $ffmpegUrl = "https://github.com/grandstyle/rist-server/releases/download/v1.0.1/ffmpeg.exe"
    Invoke-WebRequest -Uri $ffmpegUrl -OutFile "assets\ffmpeg.exe" -UseBasicParsing
}

# 1. Main App Resource compilation (.syso with Icon & Manifest)
if (Test-Path "winres\winres.json") {
    Write-Host "[1/4] Generating RIST-Server Resources & Icon..." -ForegroundColor Yellow
    & "$env:USERPROFILE\go\bin\go-winres.exe" make --in "winres\winres.json"
}

# 2. Main App Go build
Write-Host "[2/4] Compiling RIST-Server.exe..." -ForegroundColor Yellow
go build -ldflags="-s -w -H=windowsgui" -trimpath -o "RIST-Server.exe" .

if (-not (Test-Path "RIST-Server.exe")) {
    Write-Error "Build failed: RIST-Server.exe not found."
    exit 1
}

# 3. Installer Resource & Embed preparation
Write-Host "[3/4] Preparing Standalone Setup..." -ForegroundColor Yellow
Copy-Item "RIST-Server.exe" "installer\RIST-Server.exe" -Force
Copy-Item "assets\app.ico" "installer\app.ico" -Force

Push-Location "installer"
try {
    if (Test-Path "winres.json") {
        & "$env:USERPROFILE\go\bin\go-winres.exe" make --in "winres.json"
    }
    Write-Host "[4/4] Compiling Setup-RIST-Server.exe..." -ForegroundColor Yellow
    go build -ldflags="-s -w -H=windowsgui" -trimpath -o "..\Setup-RIST-Server.exe" .
} finally {
    Remove-Item "RIST-Server.exe" -Force -ErrorAction SilentlyContinue
    Remove-Item "app.ico" -Force -ErrorAction SilentlyContinue
    Remove-Item "rsrc_windows_*.syso" -Force -ErrorAction SilentlyContinue
    Pop-Location
}

# Verification
Write-Host ""
Write-Host "========================================" -ForegroundColor Green
Write-Host "            BUILD SUCCESS!              " -ForegroundColor Green
Write-Host "========================================" -ForegroundColor Green

if (Test-Path "RIST-Server.exe") {
    $item = Get-Item "RIST-Server.exe"
    $sizeMb = [math]::Round($item.Length / 1MB, 2)
    Write-Host "• Portable Server: $($item.FullName) ($sizeMb MB)" -ForegroundColor Green
}

if (Test-Path "Setup-RIST-Server.exe") {
    $setupItem = Get-Item "Setup-RIST-Server.exe"
    $setupSizeMb = [math]::Round($setupItem.Length / 1MB, 2)
    Write-Host "• Standalone Setup: $($setupItem.FullName) ($setupSizeMb MB)" -ForegroundColor Green
}
