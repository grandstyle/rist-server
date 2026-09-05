Write-Host "========================================" -ForegroundColor Cyan
Write-Host " Testing RIST Receiver Stream Forwarding" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan

$ffmpeg = "d:\VS Code\RIST Server\assets\ffmpeg.exe"

# 1. Receiver Command Simulation (what RIST Server runs):
# ffmpeg.exe -hide_banner -loglevel error -i "rist://@0.0.0.0:1968?buffer=1000" -c copy -f mpegts "udp://127.0.0.1:10000?pkt_size=1316" -progress pipe:1 -stats_period 0.5

$recArgs = @(
    "-hide_banner",
    "-loglevel", "error",
    "-i", "rist://@0.0.0.0:1968?buffer=1000",
    "-c", "copy",
    "-f", "mpegts",
    "udp://127.0.0.1:10000?pkt_size=1316",
    "-progress", "pipe:1",
    "-stats_period", "0.5"
)

$psi = New-Object System.Diagnostics.ProcessStartInfo
$psi.FileName = $ffmpeg
$psi.Arguments = $recArgs -join " "
$psi.UseShellExecute = $false
$psi.RedirectStandardOutput = $true
$psi.RedirectStandardError = $true
$psi.CreateNoWindow = $true

Write-Host "Starting Receiver process..." -ForegroundColor Yellow
$receiverProc = [System.Diagnostics.Process]::Start($psi)

Start-Sleep -Seconds 1

# 2. Start Test Sender (sending test pattern to rist://127.0.0.1:1968)
$senderArgs = @(
    "-hide_banner",
    "-loglevel", "error",
    "-re",
    "-f", "lavfi",
    "-i", "testsrc=size=640x360:rate=30",
    "-f", "lavfi",
    "-i", "sine=frequency=1000:sample_rate=44100",
    "-c:v", "libx264",
    "-preset", "ultrafast",
    "-b:v", "1500k",
    "-c:a", "aac",
    "-f", "mpegts",
    "-t", "5",
    "rist://127.0.0.1:1968"
)

Write-Host "Starting Sender (5s stream @ 1.5 Mbps)..." -ForegroundColor Yellow
$senderProc = Start-Process -FilePath $ffmpeg -ArgumentList ($senderArgs -join " ") -PassThru -NoNewWindow

# 3. Read telemetry output from receiver
$timeout = [System.DateTime]::Now.AddSeconds(8)
$linesReceived = 0

while ([System.DateTime]::Now -lt $timeout -and !$senderProc.HasExited) {
    if (!$receiverProc.StandardOutput.EndOfStream) {
        $line = $receiverProc.StandardOutput.ReadLine()
        if ($line -like "*bitrate*" -or $line -like "*total_size*" -or $line -like "*speed*") {
            Write-Host "[Telemetry Line] $line" -ForegroundColor Green
            $linesReceived++
        }
    }
    Start-Sleep -Milliseconds 100
}

Write-Host "Waiting for sender to finish..."
$senderProc.WaitForExit()

Write-Host "Terminating receiver..."
$receiverProc.Kill()
$receiverProc.WaitForExit()

if ($linesReceived -gt 0) {
    Write-Host "SUCCESS: Telemetry and RIST Stream received and parsed correctly!" -ForegroundColor Green
} else {
    Write-Error "No telemetry lines captured."
}
