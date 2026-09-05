# Quick 1-Click Setup Script for RIST Server
$ErrorActionPreference = "Stop"

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Definition
$rootDir = Split-Path -Parent $scriptDir
$srcExe = Join-Path $rootDir "RIST-Server.exe"

if (-not (Test-Path $srcExe)) {
    $srcExe = Join-Path $scriptDir "RIST-Server.exe"
}

if (-not (Test-Path $srcExe)) {
    Write-Error "RIST-Server.exe wurde nicht gefunden! Bitte zuerst build.ps1 ausführen."
    exit 1
}

$targetDir = Join-Path $env:LOCALAPPDATA "Programs\RIST Server"
Write-Host "Installiere RIST Server nach: $targetDir" -ForegroundColor Cyan
New-Item -ItemType Directory -Path $targetDir -Force | Out-Null

$targetExe = Join-Path $targetDir "RIST-Server.exe"
Copy-Item -Path $srcExe -Destination $targetExe -Force

# Desktop Shortcut
$desktopPath = Join-Path ([Environment]::GetFolderPath("Desktop")) "RIST Server.lnk"
$ws = New-Object -ComObject WScript.Shell
$s = $ws.CreateShortcut($desktopPath)
$s.TargetPath = $targetExe
$s.WorkingDirectory = $targetDir
$s.IconLocation = "$targetExe,0"
$s.Description = "RIST Receiver & Relay Server"
$s.Save()

# Start Menu Shortcut
$startDir = Join-Path $env:APPDATA "Microsoft\Windows\Start Menu\Programs\RIST Server"
New-Item -ItemType Directory -Path $startDir -Force | Out-Null
$startLnk = Join-Path $startDir "RIST Server.lnk"
$s2 = $ws.CreateShortcut($startLnk)
$s2.TargetPath = $targetExe
$s2.WorkingDirectory = $targetDir
$s2.IconLocation = "$targetExe,0"
$s2.Description = "RIST Receiver & Relay Server"
$s2.Save()

Write-Host "✔ RIST Server erfolgreich installiert mit Desktop-Verknüpfung!" -ForegroundColor Green
Write-Host "Desktop-Verknüpfung: $desktopPath" -ForegroundColor Yellow

$ans = Read-Host "Möchtest du RIST Server jetzt starten? (J/N)"
if ($ans -match "^[jJyY]") {
    Start-Process -FilePath $targetExe -WorkingDirectory $targetDir
}
