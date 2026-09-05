package telemetry

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"sync"

	"rist-server/assets"
)

type HTTPServer struct {
	parser    *Parser
	port      int
	ctrlMu    sync.RWMutex
	onToggle  func() (bool, error)
	onStart   func() error
	onStop    func()
	isRunning func() bool
	getConfig func() map[string]interface{}
	setConfig func(map[string]interface{}) error
}

func (s *HTTPServer) SetController(
	onToggle func() (bool, error),
	onStart func() error,
	onStop func(),
	isRunning func() bool,
	getConfig func() map[string]interface{},
	setConfig func(map[string]interface{}) error,
) {
	s.ctrlMu.Lock()
	defer s.ctrlMu.Unlock()
	s.onToggle = onToggle
	s.onStart = onStart
	s.onStop = onStop
	s.isRunning = isRunning
	s.getConfig = getConfig
	s.setConfig = setConfig
}

func GetLocalIP() string {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err == nil {
		defer conn.Close()
		localAddr := conn.LocalAddr().(*net.UDPAddr)
		if ip := localAddr.IP.To4(); ip != nil && !ip.IsLoopback() {
			return ip.String()
		}
	}

	addrs, err := net.InterfaceAddrs()
	if err == nil {
		for _, addr := range addrs {
			if ipnet, ok := addr.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
				if ip4 := ipnet.IP.To4(); ip4 != nil {
					return ip4.String()
				}
			}
		}
	}
	return "127.0.0.1"
}

func StartHTTPServer(parser *Parser, defaultPort int) (*HTTPServer, error) {
	s := &HTTPServer{
		parser: parser,
	}

	mux := http.NewServeMux()

	// 1. Unified JSON Stats (/stats & /api/stats)
	statsHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		stats := parser.GetStats()
		statusStr := "offline"
		bitrate := int64(stats.BitrateKbps)
		bitrateKbps := stats.BitrateKbps
		fpsNum, _ := strconv.ParseFloat(stats.Fps, 64)
		speedStr := stats.Speed

		if stats.State == StateStreaming && stats.BitrateKbps > 0 {
			statusStr = "online"
		} else if stats.State == StateListening {
			statusStr = "listening"
			bitrate = 0
			bitrateKbps = 0
			fpsNum = 0
			speedStr = "0.0x"
		} else {
			bitrate = 0
			bitrateKbps = 0
			fpsNum = 0
			speedStr = "0.0x"
		}

		s.ctrlMu.RLock()
		running := false
		if s.isRunning != nil {
			running = s.isRunning()
		}
		s.ctrlMu.RUnlock()

		resp := map[string]interface{}{
			"server_running": running,
			"status":         statusStr,
			"bitrate":        bitrate,
			"bitrate_kbps":   bitrateKbps,
			"bitrate_bps":    bitrate * 1000,
			"total_bytes":    stats.TotalBytes,
			"total_mb":       float64(stats.TotalBytes) / (1024 * 1024),
			"fps":            fpsNum,
			"speed":          speedStr,
			"duration":       stats.Duration,
		}

		_ = json.NewEncoder(w).Encode(resp)
	}

	mux.HandleFunc("/stats", statsHandler)
	mux.HandleFunc("/api/stats", statsHandler)

	// 2. Plain Text Bitrate (/bitrate)
	mux.HandleFunc("/bitrate", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		stats := parser.GetStats()
		if stats.State == StateStreaming && stats.BitrateKbps > 0 {
			_, _ = fmt.Fprintf(w, "%d\n", int64(stats.BitrateKbps))
		} else {
			_, _ = fmt.Fprintln(w, "0")
		}
	})

	// 3. SLS (SRT Live Server) Format (/sls)
	mux.HandleFunc("/sls", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		stats := parser.GetStats()
		status := "idle"
		var bitrate int64 = 0
		if stats.State == StateStreaming && stats.BitrateKbps > 0 {
			status = "publishing"
			bitrate = int64(stats.BitrateKbps)
		}

		resp := map[string]interface{}{
			"publishers": []map[string]interface{}{
				{
					"bitrate": bitrate,
					"bytes":   stats.TotalBytes,
					"status":  status,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// 4. Belabox Format (/belabox)
	mux.HandleFunc("/belabox", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		stats := parser.GetStats()
		status := "offline"
		var bitrate int64 = 0
		if stats.State == StateStreaming && stats.BitrateKbps > 0 {
			status = "online"
			bitrate = int64(stats.BitrateKbps)
		}

		resp := map[string]interface{}{
			"bitrate":     bitrate,
			"total_bytes": stats.TotalBytes,
			"status":      status,
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// 5. OBS Browser Source Overlay (/overlay)
	mux.HandleFunc("/overlay", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(htmlOverlay))
	})

	// 6. Mobile Remote Control (optimiert für Samsung Galaxy S10e) (/remote & /)
	mux.HandleFunc("/remote", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(htmlRemote))
	})

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/remote", http.StatusFound)
			return
		}
		http.NotFound(w, r)
	})

	// PWA Manifest & Service Worker
	manifestHandler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/manifest+json; charset=utf-8")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		_, _ = w.Write([]byte(manifestJSON))
	}
	mux.HandleFunc("/manifest.json", manifestHandler)
	mux.HandleFunc("/manifest.webmanifest", manifestHandler)

	mux.HandleFunc("/sw.js", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
		w.Header().Set("Service-Worker-Allowed", "/")
		_, _ = w.Write([]byte(serviceWorkerJS))
	})

	// PWA Icons
	mux.HandleFunc("/icon-192.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(assets.Icon192)
	})

	mux.HandleFunc("/icon-512.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(assets.Icon512)
	})

	mux.HandleFunc("/icon-maskable-192.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(assets.IconMaskable192)
	})

	mux.HandleFunc("/icon-maskable-512.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		_, _ = w.Write(assets.IconMaskable512)
	})

	mux.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		_, _ = w.Write(assets.AppIcon)
	})

	// 7. Remote API Controls
	mux.HandleFunc("/api/server/toggle", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		s.ctrlMu.RLock()
		toggleFn := s.onToggle
		s.ctrlMu.RUnlock()

		if toggleFn == nil {
			http.Error(w, `{"error":"controller not configured"}`, http.StatusServiceUnavailable)
			return
		}
		running, err := toggleFn()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error(), "running": running})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "running": running})
	})

	mux.HandleFunc("/api/server/start", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		s.ctrlMu.RLock()
		startFn := s.onStart
		s.ctrlMu.RUnlock()

		if startFn == nil {
			http.Error(w, `{"error":"controller not configured"}`, http.StatusServiceUnavailable)
			return
		}
		if err := startFn(); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "running": true})
	})

	mux.HandleFunc("/api/server/stop", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		s.ctrlMu.RLock()
		stopFn := s.onStop
		s.ctrlMu.RUnlock()

		if stopFn == nil {
			http.Error(w, `{"error":"controller not configured"}`, http.StatusServiceUnavailable)
			return
		}
		stopFn()
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "running": false})
	})

	mux.HandleFunc("/api/config", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		s.ctrlMu.RLock()
		getFn := s.getConfig
		setFn := s.setConfig
		s.ctrlMu.RUnlock()

		if r.Method == http.MethodPost {
			if setFn == nil {
				http.Error(w, `{"error":"controller not configured"}`, http.StatusServiceUnavailable)
				return
			}
			var payload map[string]interface{}
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				http.Error(w, `{"error":"invalid json"}`, http.StatusBadRequest)
				return
			}
			if err := setFn(payload); err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": false, "error": err.Error()})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			return
		}

		if getFn != nil {
			_ = json.NewEncoder(w).Encode(getFn())
		} else {
			_ = json.NewEncoder(w).Encode(map[string]interface{}{})
		}
	})

	// Auf allen Interfaces (0.0.0.0) lauschen für LAN / Netzwerk-Zugriff
	var listener net.Listener
	var err error
	port := defaultPort

	for p := defaultPort; p < defaultPort+10; p++ {
		addr := fmt.Sprintf("0.0.0.0:%d", p)
		listener, err = net.Listen("tcp", addr)
		if err == nil {
			port = p
			break
		}
	}

	if listener == nil {
		return nil, fmt.Errorf("failed to bind HTTP server on ports %d-%d: %w", defaultPort, defaultPort+9, err)
	}

	s.port = port

	server := &http.Server{
		Handler: mux,
	}

	go func() {
		_ = server.Serve(listener)
	}()

	return s, nil
}

func (s *HTTPServer) GetPort() int {
	return s.port
}
