package telemetry

import (
	"fmt"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ServerState int

const (
	StateOffline ServerState = iota
	StateListening
	StateStreaming
)

func (s ServerState) String() string {
	switch s {
	case StateOffline:
		return "OFFLINE"
	case StateListening:
		return "WARTE AUF STREAM..."
	case StateStreaming:
		return "LIVE STREAMING"
	default:
		return "UNKNOWN"
	}
}

type Stats struct {
	State          ServerState
	BitrateStr     string  // z.B. "4520.1 kbit/s" oder "4.52 Mbit/s"
	BitrateKbps    float64 // Numerischer Wert in kbps
	TotalBytes     int64   // Gesamt übertragene Bytes
	TotalFormatted string  // z.B. "124.50 MB"
	Duration       string  // z.B. "00:05:12"
	Speed          string  // z.B. "1.00x"
	Fps            string  // z.B. "60.0"
	LastPacketTime time.Time
}

type Parser struct {
	mu             sync.RWMutex
	stats          Stats
	onUpdate       func(Stats)
	lastBytes      int64
	lastPacketTime time.Time
	hasError       bool
}

func NewParser(onUpdate func(Stats)) *Parser {
	p := &Parser{
		stats: Stats{
			State:          StateOffline,
			BitrateStr:     "0.0 kbit/s",
			TotalFormatted: "0.00 MB",
			Duration:       "00:00:00",
			Speed:          "0.0x",
			Fps:            "0.0",
		},
		onUpdate: onUpdate,
	}

	// Watchdog: Erkennt in Echtzeit, wenn kein Stream mehr gesendet wird oder ein Fehler auftritt
	go func() {
		ticker := time.NewTicker(250 * time.Millisecond)
		defer ticker.Stop()

		for range ticker.C {
			p.mu.Lock()
			if p.stats.State == StateStreaming {
				timedOut := !p.lastPacketTime.IsZero() && time.Since(p.lastPacketTime) > 1200*time.Millisecond
				if timedOut || p.hasError {
					// Stream abgerissen / beendet / Fehler -> Sofort zurücksetzen
					p.stats.State = StateListening
					p.stats.BitrateStr = "0.0 kbit/s"
					p.stats.BitrateKbps = 0
					p.stats.Fps = "0.0"
					p.stats.Speed = "0.0x"
					s := p.stats
					onUp := p.onUpdate
					p.mu.Unlock()

					if onUp != nil {
						onUp(s)
					}
					continue
				}
			}
			p.mu.Unlock()
		}
	}()

	return p
}

func (p *Parser) SetOnUpdate(onUpdate func(Stats)) {
	p.mu.Lock()
	p.onUpdate = onUpdate
	p.mu.Unlock()
}

// HandleStreamError setzt die Bitrate und den Status sofort auf 0 / Listening zurück
func (p *Parser) HandleStreamError(errText string) {
	p.mu.Lock()
	p.hasError = true
	p.stats.BitrateStr = "0.0 kbit/s"
	p.stats.BitrateKbps = 0
	p.stats.Fps = "0.0"
	p.stats.Speed = "0.0x"
	if p.stats.State == StateStreaming {
		p.stats.State = StateListening
	}
	s := p.stats
	onUp := p.onUpdate
	p.mu.Unlock()

	if onUp != nil {
		onUp(s)
	}
}

func (p *Parser) SetState(state ServerState) {
	p.mu.Lock()
	p.stats.State = state
	if state == StateOffline {
		p.stats.BitrateStr = "0.0 kbit/s"
		p.stats.BitrateKbps = 0
		p.stats.TotalBytes = 0
		p.stats.TotalFormatted = "0.00 MB"
		p.stats.Duration = "00:00:00"
		p.stats.Speed = "0.0x"
		p.stats.Fps = "0.0"
		p.stats.LastPacketTime = time.Time{}
		p.lastPacketTime = time.Time{}
		p.lastBytes = 0
		p.hasError = false
	} else if state == StateListening {
		p.stats.BitrateStr = "0.0 kbit/s"
		p.stats.BitrateKbps = 0
		p.stats.Fps = "0.0"
		p.stats.Speed = "0.0x"
		p.stats.LastPacketTime = time.Time{}
		p.lastPacketTime = time.Time{}
		p.hasError = false
	}
	s := p.stats
	onUp := p.onUpdate
	p.mu.Unlock()

	if onUp != nil {
		onUp(s)
	}
}

func (p *Parser) GetStats() Stats {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return p.stats
}

// ProcessLine verarbeitet eine Zeile aus FFmpegs -progress pipe:1
func (p *Parser) ProcessLine(line string) {
	line = strings.TrimSpace(line)
	if line == "" {
		return
	}

	parts := strings.SplitN(line, "=", 2)
	if len(parts) != 2 {
		return
	}

	key := strings.TrimSpace(parts[0])
	val := strings.TrimSpace(parts[1])

	p.mu.Lock()
	defer p.mu.Unlock()

	switch key {
	case "bitrate":
		if val != "N/A" && val != "" {
			cleanVal := strings.TrimSuffix(val, "kbits/s")
			cleanVal = strings.TrimSuffix(cleanVal, "kbit/s")
			cleanVal = strings.TrimSpace(cleanVal)
			if kbps, err := strconv.ParseFloat(cleanVal, 64); err == nil {
				if kbps <= 0 {
					p.stats.BitrateKbps = 0
					p.stats.BitrateStr = "0.0 kbit/s"
					if p.stats.State == StateStreaming {
						p.stats.State = StateListening
					}
				} else {
					// Nur übernehmen, wenn kein Error vorliegt und vor kurzem tatsächlich Daten empfangen wurden
					isRecent := !p.lastPacketTime.IsZero() && time.Since(p.lastPacketTime) <= 1200*time.Millisecond
					if !p.hasError && (isRecent || p.stats.TotalBytes == 0) {
						p.stats.BitrateKbps = kbps
						if kbps >= 1000.0 {
							p.stats.BitrateStr = fmt.Sprintf("%.2f Mbit/s", kbps/1000.0)
						} else {
							p.stats.BitrateStr = fmt.Sprintf("%.1f kbit/s", kbps)
						}
						p.stats.State = StateStreaming
					} else {
						// Falls Stream abgerissen oder Error: Bitrate auf 0
						p.stats.BitrateKbps = 0
						p.stats.BitrateStr = "0.0 kbit/s"
						if p.stats.State == StateStreaming {
							p.stats.State = StateListening
						}
					}
				}
			}
		} else if val == "N/A" {
			p.stats.BitrateKbps = 0
			p.stats.BitrateStr = "0.0 kbit/s"
			if p.stats.State == StateStreaming {
				p.stats.State = StateListening
			}
		}

	case "total_size":
		if val != "N/A" && val != "" {
			if b, err := strconv.ParseInt(val, 10, 64); err == nil {
				// Wenn Bytes wachsen, ist der Stream aktiv und liefert neue Daten
				if b > p.lastBytes && p.stats.State != StateOffline {
					p.stats.State = StateStreaming
					p.lastPacketTime = time.Now()
					p.stats.LastPacketTime = time.Now()
					p.hasError = false
					p.lastBytes = b
				}
				p.stats.TotalBytes = b
				p.stats.TotalFormatted = FormatBytes(b)
			}
		}

	case "out_time":
		if val != "N/A" && val != "" {
			dotIdx := strings.Index(val, ".")
			if dotIdx != -1 {
				p.stats.Duration = val[:dotIdx]
			} else {
				p.stats.Duration = val
			}
		}

	case "speed":
		if val != "N/A" && val != "" {
			if p.stats.State == StateStreaming && !p.hasError && (!p.lastPacketTime.IsZero() && time.Since(p.lastPacketTime) <= 1200*time.Millisecond) {
				p.stats.Speed = val
			} else {
				p.stats.Speed = "0.0x"
			}
		}

	case "fps":
		if val != "N/A" && val != "" {
			if p.stats.State == StateStreaming && !p.hasError && (!p.lastPacketTime.IsZero() && time.Since(p.lastPacketTime) <= 1200*time.Millisecond) {
				p.stats.Fps = val
			} else {
				p.stats.Fps = "0.0"
			}
		}

	case "progress":
		if val == "end" {
			p.stats.BitrateKbps = 0
			p.stats.BitrateStr = "0.0 kbit/s"
			p.stats.Fps = "0.0"
			p.stats.Speed = "0.0x"
			if p.stats.State == StateStreaming {
				p.stats.State = StateListening
			}
		}
		s := p.stats
		onUp := p.onUpdate
		if onUp != nil {
			go onUp(s)
		}
	}
}

func FormatBytes(bytes int64) string {
	const (
		kb = 1024
		mb = 1024 * kb
		gb = 1024 * mb
	)
	switch {
	case bytes >= gb:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(gb))
	case bytes >= mb:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(mb))
	case bytes >= kb:
		return fmt.Sprintf("%.1f KB", float64(bytes)/float64(kb))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}
