package engine

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime/debug"
	"strings"
	"sync"
	"syscall"
	"unsafe"

	"rist-server/telemetry"
)

var (
	kernel32                     = syscall.NewLazyDLL("kernel32.dll")
	procCreateJobObjectW         = kernel32.NewProc("CreateJobObjectW")
	procSetInformationJobObject  = kernel32.NewProc("SetInformationJobObject")
	procAssignProcessToJobObject = kernel32.NewProc("AssignProcessToJobObject")
	procCloseHandle              = kernel32.NewProc("CloseHandle")
)

const (
	jobObjectExtendedLimitInformation = 9
	jobObjectLimitKillOnJobClose      = 0x2000
	createNoWindow                    = 0x08000000
)

type ioCounters struct {
	ReadOperationCount  uint64
	WriteOperationCount uint64
	OtherOperationCount uint64
	ReadTransferCount   uint64
	WriteTransferCount  uint64
	OtherTransferCount  uint64
}

type jobObjectBasicLimitInformation struct {
	PerProcessUserTimeLimit int64
	PerJobUserTimeLimit     int64
	LimitFlags              uint32
	MinimumWorkingSetSize   uintptr
	MaximumWorkingSetSize   uintptr
	ActiveProcessLimit      uint32
	Affinity                uintptr
	PriorityClass           uint32
	SchedulingClass         uint32
}

type jobObjectExtendedLimitInfo struct {
	BasicLimitInformation jobObjectBasicLimitInformation
	IoInfo                ioCounters
	ProcessMemoryLimit    uintptr
	JobMemoryLimit        uintptr
	PeakProcessMemoryUsed uintptr
	PeakJobMemoryUsed     uintptr
}

type ServerConfig struct {
	RistPort    int
	OutHost     string
	OutPort     int
	BufferMs    int
	PacketSize  int
	StatsPeriod float64
}

func DefaultConfig() ServerConfig {
	return ServerConfig{
		RistPort:    1968,
		OutHost:     "127.0.0.1",
		OutPort:     10000,
		BufferMs:    1000,
		PacketSize:  1316,
		StatsPeriod: 0.5,
	}
}

type Engine struct {
	mu           sync.Mutex
	config       ServerConfig
	cmd          *exec.Cmd
	jobHandle    syscall.Handle
	stdinPipe    io.WriteCloser
	telemetry    *telemetry.Parser
	ffmpegPath   string
	isRunning    bool
	embeddedData []byte
	onLog        func(string)
}

func NewEngine(embeddedFFmpeg []byte, tel *telemetry.Parser, onLog func(string)) *Engine {
	return &Engine{
		config:       DefaultConfig(),
		telemetry:    tel,
		embeddedData: embeddedFFmpeg,
		onLog:        onLog,
	}
}

func (e *Engine) log(msg string) {
	if e.onLog != nil {
		e.onLog(msg)
	}
}

// EnsureFFmpeg extrahiert die eingebettete ffmpeg.exe ins Temp-Verzeichnis
func (e *Engine) EnsureFFmpeg() (string, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.ffmpegPath != "" {
		if _, err := os.Stat(e.ffmpegPath); err == nil {
			return e.ffmpegPath, nil
		}
	}

	tempDir := filepath.Join(os.TempDir(), "rist_receiver_engine")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", fmt.Errorf("temp dir creation error: %w", err)
	}

	targetPath := filepath.Join(tempDir, "ffmpeg.exe")

	// Prüfe, ob Datei existiert und identisch ist
	if fi, err := os.Stat(targetPath); err == nil && fi.Size() == int64(len(e.embeddedData)) {
		e.ffmpegPath = targetPath
		return targetPath, nil
	}

	if len(e.embeddedData) == 0 {
		return "", fmt.Errorf("no embedded FFmpeg binary found")
	}

	hash := sha256.Sum256(e.embeddedData)
	e.log(fmt.Sprintf("Extracting FFmpeg engine (SHA256: %s...)", hex.EncodeToString(hash[:4])))

	tmpTarget := targetPath + ".tmp"
	if err := os.WriteFile(tmpTarget, e.embeddedData, 0755); err != nil {
		return "", fmt.Errorf("failed to write ffmpeg temp file: %w", err)
	}

	_ = os.Remove(targetPath)
	if err := os.Rename(tmpTarget, targetPath); err != nil {
		// Fallback: versuche direkt zu schreiben
		if err2 := os.WriteFile(targetPath, e.embeddedData, 0755); err2 != nil {
			return "", fmt.Errorf("failed to place ffmpeg.exe: %w", err2)
		}
	}

	e.ffmpegPath = targetPath
	e.log("FFmpeg engine extracted successfully.")
	// Freigabe des Embed-Speichers für minimalen RAM-Footprint
	e.embeddedData = nil
	debug.FreeOSMemory()
	return targetPath, nil
}

// Start startet den RIST-Forwarding-Prozess
func (e *Engine) Start(cfg ServerConfig) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	if e.isRunning {
		return fmt.Errorf("server is already running")
	}

	e.config = cfg

	ffmpegPath, err := e.ensureFFmpegLocked()
	if err != nil {
		return fmt.Errorf("ffmpeg preparation error: %w", err)
	}

	// Windows Job Object erstellen für zuverlässiges Killen bei Exit
	job, err := createKillOnCloseJob()
	if err != nil {
		e.log(fmt.Sprintf("Warning: could not create Windows Job Object: %v", err))
	} else {
		e.jobHandle = job
	}

	// ffmpeg.exe -hide_banner -loglevel error -i "rist://@0.0.0.0:[PORT]?buffer=1000" -c copy -bsf:v dump_extra -mpegts_flags +resend_headers -f mpegts "udp://[HOST]:[PORT]?pkt_size=1316&buffer_size=1048576" -progress pipe:1 -stats_period 0.5
	outHost := strings.TrimSpace(cfg.OutHost)
	if outHost == "" {
		outHost = "127.0.0.1"
	}
	broadcastOpt := ""
	if outHost == "255.255.255.255" || strings.HasSuffix(outHost, ".255") {
		broadcastOpt = "&broadcast=1"
	}

	inURL := fmt.Sprintf("rist://@0.0.0.0:%d?buffer=%d", cfg.RistPort, cfg.BufferMs)
	outURL := fmt.Sprintf("udp://%s:%d?pkt_size=%d&buffer_size=1048576%s", outHost, cfg.OutPort, cfg.PacketSize, broadcastOpt)
	statsPeriodStr := fmt.Sprintf("%.2f", cfg.StatsPeriod)

	args := []string{
		"-hide_banner",
		"-loglevel", "error",
		"-i", inURL,
		"-c", "copy",
		"-bsf:v", "dump_extra",
		"-mpegts_flags", "+resend_headers",
		"-f", "mpegts",
		outURL,
		"-progress", "pipe:1",
		"-stats_period", statsPeriodStr,
	}

	cmd := exec.Command(ffmpegPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: createNoWindow,
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		e.cleanupJobLocked()
		return fmt.Errorf("failed to open stdout pipe: %w", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		e.cleanupJobLocked()
		return fmt.Errorf("failed to open stderr pipe: %w", err)
	}

	stdin, err := cmd.StdinPipe()
	if err != nil {
		e.cleanupJobLocked()
		return fmt.Errorf("failed to open stdin pipe: %w", err)
	}
	e.stdinPipe = stdin

	if err := cmd.Start(); err != nil {
		e.cleanupJobLocked()
		return fmt.Errorf("failed to start ffmpeg: %w", err)
	}

	if e.jobHandle != 0 && cmd.Process != nil {
		assignProcessToJob(e.jobHandle, cmd.Process.Pid)
	}

	e.cmd = cmd
	e.isRunning = true
	e.telemetry.SetState(telemetry.StateListening)
	e.log(fmt.Sprintf("Started RIST Receiver on :%d -> UDP %s:%d", cfg.RistPort, outHost, cfg.OutPort))

	// Stdout Reader für Telemetrie
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			e.telemetry.ProcessLine(line)
		}
	}()

	// Stderr Reader für Fehlermeldungen
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			errText := scanner.Text()
			if errText != "" {
				e.log(fmt.Sprintf("[FFmpeg Error] %s", errText))
				if e.telemetry != nil {
					e.telemetry.HandleStreamError(errText)
				}
			}
		}
	}()

	// Watchdog Goroutine für Prozessende
	go func() {
		_ = cmd.Wait()
		e.mu.Lock()
		e.isRunning = false
		e.cmd = nil
		e.stdinPipe = nil
		e.cleanupJobLocked()
		e.telemetry.SetState(telemetry.StateOffline)
		e.mu.Unlock()
		e.log("RIST Receiver stopped.")
	}()

	return nil
}

func (e *Engine) ensureFFmpegLocked() (string, error) {
	if e.ffmpegPath != "" {
		if _, err := os.Stat(e.ffmpegPath); err == nil {
			return e.ffmpegPath, nil
		}
	}

	tempDir := filepath.Join(os.TempDir(), "rist_receiver_engine")
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return "", err
	}

	targetPath := filepath.Join(tempDir, "ffmpeg.exe")
	if fi, err := os.Stat(targetPath); err == nil && fi.Size() == int64(len(e.embeddedData)) {
		e.ffmpegPath = targetPath
		return targetPath, nil
	}

	if len(e.embeddedData) == 0 {
		return "", fmt.Errorf("embedded ffmpeg is empty")
	}

	_ = os.WriteFile(targetPath, e.embeddedData, 0755)
	e.ffmpegPath = targetPath
	return targetPath, nil
}

// Stop beendet den FFmpeg-Prozess sauber
func (e *Engine) Stop() {
	e.mu.Lock()
	defer e.mu.Unlock()

	if !e.isRunning || e.cmd == nil {
		return
	}

	e.log("Stopping RIST Receiver...")

	// 1. Versuche sauberes Quit über stdin
	if e.stdinPipe != nil {
		_, _ = e.stdinPipe.Write([]byte("q\n"))
		_ = e.stdinPipe.Close()
	}

	// 2. Kill nach kurzer Karenzzeit falls nötig
	if e.cmd != nil && e.cmd.Process != nil {
		_ = e.cmd.Process.Kill()
	}

	e.cleanupJobLocked()
	e.isRunning = false
	e.telemetry.SetState(telemetry.StateOffline)
}

func (e *Engine) IsRunning() bool {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.isRunning
}

func (e *Engine) GetConfig() ServerConfig {
	e.mu.Lock()
	defer e.mu.Unlock()
	return e.config
}

func (e *Engine) cleanupJobLocked() {
	if e.jobHandle != 0 {
		_, _, _ = procCloseHandle.Call(uintptr(e.jobHandle))
		e.jobHandle = 0
	}
}

func createKillOnCloseJob() (syscall.Handle, error) {
	r1, _, err := procCreateJobObjectW.Call(0, 0)
	if r1 == 0 {
		return 0, err
	}
	job := syscall.Handle(r1)

	var info jobObjectExtendedLimitInfo
	info.BasicLimitInformation.LimitFlags = jobObjectLimitKillOnJobClose

	r2, _, err2 := procSetInformationJobObject.Call(
		r1,
		jobObjectExtendedLimitInformation,
		uintptr(unsafe.Pointer(&info)),
		uintptr(unsafe.Sizeof(info)),
	)
	if r2 == 0 {
		_, _, _ = procCloseHandle.Call(r1)
		return 0, err2
	}

	return job, nil
}

const (
	processSetQuota  = 0x0100
	processTerminate = 0x0001
)

func assignProcessToJob(job syscall.Handle, pid int) {
	hProc, err := syscall.OpenProcess(processSetQuota|processTerminate, false, uint32(pid))
	if err != nil {
		return
	}
	defer syscall.CloseHandle(hProc)

	_, _, _ = procAssignProcessToJobObject.Call(uintptr(job), uintptr(hProc))
}
