package telemetry

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

func TestStreamErrorDropsBitrateToZero(t *testing.T) {
	p := NewParser(nil)
	p.SetState(StateListening)

	// 1. Simuliere eingehenden Stream
	p.ProcessLine("total_size=100000")
	p.ProcessLine("bitrate= 4500.0kbits/s")
	p.ProcessLine("fps=60.0")
	p.ProcessLine("speed=1.0x")
	p.ProcessLine("progress=continue")

	stats := p.GetStats()
	if stats.State != StateStreaming {
		t.Fatalf("Expected StateStreaming, got %v", stats.State)
	}
	if stats.BitrateKbps != 4500.0 {
		t.Fatalf("Expected BitrateKbps 4500.0, got %f", stats.BitrateKbps)
	}

	// 2. Simuliere Stream-Fehler
	p.HandleStreamError("Connection reset by peer")

	stats = p.GetStats()
	if stats.State != StateListening {
		t.Fatalf("Expected StateListening after error, got %v", stats.State)
	}
	if stats.BitrateKbps != 0 {
		t.Fatalf("Expected BitrateKbps 0 after error, got %f", stats.BitrateKbps)
	}
	if stats.BitrateStr != "0.0 kbit/s" {
		t.Fatalf("Expected '0.0 kbit/s' after error, got %s", stats.BitrateStr)
	}
	if stats.Fps != "0.0" {
		t.Fatalf("Expected '0.0' fps after error, got %s", stats.Fps)
	}

	// 3. Stale FFmpeg lines ohne neue Bytes dürfen Bitrate nicht wieder erhöhen
	p.ProcessLine("bitrate= 4200.0kbits/s")
	p.ProcessLine("total_size=100000") // Gleiche Bytes!
	p.ProcessLine("progress=continue")

	stats = p.GetStats()
	if stats.BitrateKbps != 0 {
		t.Fatalf("Expected BitrateKbps to remain 0 on stale bytes, got %f", stats.BitrateKbps)
	}
	if stats.State != StateListening {
		t.Fatalf("Expected StateListening to remain on stale bytes, got %v", stats.State)
	}

	// 4. Neuer Stream kommt rein -> Bytes wachsen wieder
	p.ProcessLine("total_size=200000")
	p.ProcessLine("bitrate= 5000.0kbits/s")
	p.ProcessLine("progress=continue")

	stats = p.GetStats()
	if stats.State != StateStreaming {
		t.Fatalf("Expected StateStreaming after recovery, got %v", stats.State)
	}
	if stats.BitrateKbps != 5000.0 {
		t.Fatalf("Expected BitrateKbps 5000.0 after recovery, got %f", stats.BitrateKbps)
	}
}

func TestWatchdogZeroesBitrateWhenStreamStops(t *testing.T) {
	p := NewParser(nil)
	p.SetState(StateListening)

	// Stream startet
	p.ProcessLine("total_size=50000")
	p.ProcessLine("bitrate= 3000.0kbits/s")
	p.ProcessLine("progress=continue")

	stats := p.GetStats()
	if stats.State != StateStreaming || stats.BitrateKbps != 3000.0 {
		t.Fatalf("Stream should be streaming at 3000 kbps, got state=%v, kbps=%f", stats.State, stats.BitrateKbps)
	}

	// Keine neuen Bytes für 1.4 Sekunden
	time.Sleep(1400 * time.Millisecond)

	stats = p.GetStats()
	if stats.State != StateListening {
		t.Fatalf("Expected StateListening after watchdog timeout, got %v", stats.State)
	}
	if stats.BitrateKbps != 0 {
		t.Fatalf("Expected BitrateKbps 0 after watchdog timeout, got %f", stats.BitrateKbps)
	}
	if stats.BitrateStr != "0.0 kbit/s" {
		t.Fatalf("Expected '0.0 kbit/s', got %s", stats.BitrateStr)
	}
}

func TestHTTPStatsEndpointsReturnZeroWhenNotStreaming(t *testing.T) {
	p := NewParser(nil)
	p.SetState(StateListening)

	srv, err := StartHTTPServer(p, 19686)
	if err != nil {
		t.Fatalf("Failed to start test HTTP server: %v", err)
	}
	port := srv.GetPort()

	// 1. Check /stats
	resp, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/stats", port))
	if err != nil {
		t.Fatalf("GET /stats failed: %v", err)
	}
	defer resp.Body.Close()

	var data map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		t.Fatalf("Failed to decode /stats json: %v", err)
	}

	if data["status"] != "listening" {
		t.Errorf("Expected status 'listening', got %v", data["status"])
	}
	if data["bitrate"] != float64(0) {
		t.Errorf("Expected bitrate 0, got %v", data["bitrate"])
	}
	if data["bitrate_kbps"] != float64(0) {
		t.Errorf("Expected bitrate_kbps 0, got %v", data["bitrate_kbps"])
	}

	// 2. Check /bitrate
	respBitrate, err := http.Get(fmt.Sprintf("http://127.0.0.1:%d/bitrate", port))
	if err != nil {
		t.Fatalf("GET /bitrate failed: %v", err)
	}
	defer respBitrate.Body.Close()
	body, _ := io.ReadAll(respBitrate.Body)
	if strings.TrimSpace(string(body)) != "0" {
		t.Errorf("Expected /bitrate '0', got '%s'", strings.TrimSpace(string(body)))
	}
}

func TestPWAEndpoints(t *testing.T) {
	p := NewParser(nil)
	srv, err := StartHTTPServer(p, 18690)
	if err != nil {
		t.Fatalf("Failed to start HTTP server: %v", err)
	}
	port := srv.GetPort()

	baseURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	// 1. /manifest.json
	respManifest, err := http.Get(baseURL + "/manifest.json")
	if err != nil {
		t.Fatalf("GET /manifest.json failed: %v", err)
	}
	defer respManifest.Body.Close()
	if respManifest.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for /manifest.json, got %d", respManifest.StatusCode)
	}
	var manifestData map[string]interface{}
	if err := json.NewDecoder(respManifest.Body).Decode(&manifestData); err != nil {
		t.Fatalf("Failed to parse /manifest.json: %v", err)
	}
	if manifestData["short_name"] != "RIST Remote" {
		t.Errorf("Expected short_name 'RIST Remote', got %v", manifestData["short_name"])
	}
	if manifestData["display"] != "standalone" {
		t.Errorf("Expected display 'standalone', got %v", manifestData["display"])
	}

	// 2. /sw.js
	respSW, err := http.Get(baseURL + "/sw.js")
	if err != nil {
		t.Fatalf("GET /sw.js failed: %v", err)
	}
	defer respSW.Body.Close()
	if respSW.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 for /sw.js, got %d", respSW.StatusCode)
	}
	swBody, _ := io.ReadAll(respSW.Body)
	if !strings.Contains(string(swBody), "serviceWorker") && !strings.Contains(string(swBody), "CACHE_NAME") {
		t.Errorf("sw.js content invalid")
	}

	// 3. Icons
	for _, iconPath := range []string{"/icon-192.png", "/icon-512.png", "/icon-maskable-192.png", "/icon-maskable-512.png"} {
		respIcon, err := http.Get(baseURL + iconPath)
		if err != nil {
			t.Fatalf("GET %s failed: %v", iconPath, err)
		}
		defer respIcon.Body.Close()
		if respIcon.StatusCode != http.StatusOK {
			t.Errorf("Expected 200 for %s, got %d", iconPath, respIcon.StatusCode)
		}
		if ct := respIcon.Header.Get("Content-Type"); ct != "image/png" {
			t.Errorf("Expected Content-Type image/png for %s, got %s", iconPath, ct)
		}
		iconData, _ := io.ReadAll(respIcon.Body)
		if len(iconData) < 100 {
			t.Errorf("Icon %s too small: %d bytes", iconPath, len(iconData))
		}
	}

	// 4. /remote contains manifest link
	respRemote, err := http.Get(baseURL + "/remote")
	if err != nil {
		t.Fatalf("GET /remote failed: %v", err)
	}
	defer respRemote.Body.Close()
	remoteBody, _ := io.ReadAll(respRemote.Body)
	if !strings.Contains(string(remoteBody), "rel=\"manifest\"") {
		t.Errorf("/remote does not contain manifest link")
	}
	if !strings.Contains(string(remoteBody), "sw.js") {
		t.Errorf("/remote does not contain sw.js registration")
	}
}

