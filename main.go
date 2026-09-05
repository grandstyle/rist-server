package main

import (
	"log"

	"rist-server/assets"
	"rist-server/engine"
	"rist-server/telemetry"
	"rist-server/ui"
)

func main() {
	tel := telemetry.NewParser(nil)

	// Starte lokalen HTTP Stats Server für NOALBS / Bitrate Switcher
	httpSrv, err := telemetry.StartHTTPServer(tel, 8686)
	if err != nil {
		log.Printf("Warning: could not start HTTP stats server: %v", err)
	}

	apiPort := 8686
	if httpSrv != nil {
		apiPort = httpSrv.GetPort()
	}

	lanIP := telemetry.GetLocalIP()
	log.Printf("NOALBS Stats API active on http://%s:%d/stats", lanIP, apiPort)

	eng := engine.NewEngine(assets.FFmpegBinary, tel, func(msg string) {
		log.Println(msg)
		if globalUI := ui.GetGlobalUI(); globalUI != nil {
			globalUI.SetLog(msg)
		}
	})

	if err := ui.RunWithPortAndIP(eng, tel, apiPort, lanIP, httpSrv); err != nil {
		log.Fatalf("UI error: %v", err)
	}
}
