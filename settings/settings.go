package settings

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"golang.org/x/sys/windows/registry"
)

const (
	RegKeyRun     = `Software\Microsoft\Windows\CurrentVersion\Run`
	RegValueName  = "RISTReceiverServer"
	AppDirName    = "RIST-Server"
	ConfigFileName = "config.json"
)

type Config struct {
	RistPort         int    `json:"rist_port"`
	OutHost          string `json:"out_host"`
	OutPort          int    `json:"out_port"`
	BufferMs         int    `json:"buffer_ms"`
	AudioAlerts      bool   `json:"audio_alerts"`
	MinimizeToTray   bool   `json:"minimize_to_tray"`
	AutoStartStream  bool   `json:"auto_start_stream"`
	AutoStartWindows bool   `json:"auto_start_windows"`
}

func DefaultConfig() Config {
	return Config{
		RistPort:         1968,
		OutHost:          "127.0.0.1",
		OutPort:          10000,
		BufferMs:         1000,
		AudioAlerts:      true,
		MinimizeToTray:   true,
		AutoStartStream:  false,
		AutoStartWindows: false,
	}
}

func getConfigPath() (string, error) {
	appData, err := os.UserConfigDir()
	if err != nil {
		appData = os.TempDir()
	}
	dir := filepath.Join(appData, AppDirName)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", err
	}
	return filepath.Join(dir, ConfigFileName), nil
}

func LoadConfig() Config {
	cfg := DefaultConfig()
	cfg.AutoStartWindows = IsWindowsAutoStartEnabled()

	path, err := getConfigPath()
	if err != nil {
		return cfg
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return cfg
	}

	_ = json.Unmarshal(data, &cfg)

	// Validiere Werte
	if cfg.RistPort <= 0 || cfg.RistPort > 65535 {
		cfg.RistPort = 1968
	}
	cfg.OutHost = strings.TrimSpace(cfg.OutHost)
	if cfg.OutHost == "" {
		cfg.OutHost = "127.0.0.1"
	}
	if cfg.OutPort <= 0 || cfg.OutPort > 65535 {
		cfg.OutPort = 10000
	}
	if cfg.BufferMs < 100 || cfg.BufferMs > 10000 {
		cfg.BufferMs = 1000
	}

	// Synchronisiere Windows Autostart Status
	cfg.AutoStartWindows = IsWindowsAutoStartEnabled()

	return cfg
}

// ParseOutTarget parst Eingaben wie "10000", "192.168.1.50:10000", "192.168.1.50" oder "127.0.0.1:10000"
func ParseOutTarget(raw string, defaultPort int) (string, int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return "127.0.0.1", defaultPort, nil
	}

	// Falls nur ein Port eingegeben wurde (z.B. "10000")
	if port, err := strconv.Atoi(raw); err == nil {
		if port > 0 && port <= 65535 {
			return "127.0.0.1", port, nil
		}
		return "", 0, fmt.Errorf("Port muss zwischen 1 und 65535 liegen")
	}

	// Falls "host:port" eingegeben wurde
	if strings.Contains(raw, ":") {
		host, portStr, err := net.SplitHostPort(raw)
		if err != nil {
			return "", 0, fmt.Errorf("ungültiges Format (erwartet: IP:Port oder nur Port)")
		}
		host = strings.TrimSpace(host)
		if host == "" {
			host = "127.0.0.1"
		}
		port, err := strconv.Atoi(portStr)
		if err != nil || port <= 0 || port > 65535 {
			return "", 0, fmt.Errorf("ungültiger Port in '%s'", raw)
		}
		return host, port, nil
	}

	// Falls nur eine IP ohne Doppelpunkt angegeben wurde
	return raw, defaultPort, nil
}

// ParseRistPort parst Eingaben wie "1968", ":1968", "192.168.1.50:1968" und gibt den Port zurück.
func ParseRistPort(raw string, defaultPort int) (int, error) {
	_, port, err := ParseOutTarget(raw, defaultPort)
	if err != nil {
		return 0, err
	}
	if port <= 0 || port > 65535 {
		return 0, fmt.Errorf("Port muss zwischen 1 und 65535 liegen")
	}
	return port, nil
}

func SaveConfig(cfg Config) error {
	path, err := getConfigPath()
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func IsWindowsAutoStartEnabled() bool {
	k, err := registry.OpenKey(registry.CURRENT_USER, RegKeyRun, registry.QUERY_VALUE)
	if err != nil {
		return false
	}
	defer k.Close()

	val, _, err := k.GetStringValue(RegValueName)
	if err != nil || val == "" {
		return false
	}
	return true
}

func SetWindowsAutoStart(enable bool) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, RegKeyRun, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()

	if enable {
		exePath, err := os.Executable()
		if err != nil {
			return err
		}
		exePath, _ = filepath.Abs(exePath)
		cmd := `"` + exePath + `"`
		return k.SetStringValue(RegValueName, cmd)
	} else {
		_ = k.DeleteValue(RegValueName)
		return nil
	}
}
