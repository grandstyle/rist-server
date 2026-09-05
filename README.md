# ⚡ RIST Receiver Server

[![Release](https://img.shields.io/github/v/release/grandstyle/rist-server?style=flat-square&color=00f0ff)](https://github.com/grandstyle/rist-server/releases)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat-square&logo=go)](https://golang.org)
[![Platform](https://img.shields.io/badge/Platform-Windows%2010%20%2F%2011-blue?style=flat-square&logo=windows)](https://github.com/grandstyle/rist-server/releases)
[![License](https://img.shields.io/badge/License-MIT-green?style=flat-square)](LICENSE)
[![Support Ko-fi](https://img.shields.io/badge/Support-Ko--fi-ff5e5b?style=flat-square&logo=kofi)](https://ko-fi.com/jumpgeil)

Eine ultra-schlanke, portable Windows Single-File-Anwendung (RIST-Server.exe) in Go zum Empfang von **RIST (Reliable Internet Stream Protocol)** Video-Streams, Weiterleitung an OBS Studio / vMix / VLC via UDP und automatischer **Telemetrie HTTP-API für Bitrate-Switcher (NOALBS, Belabox, OBS Scripts)**.

---

## 🚀 Highlights & Features

- ⚡ **Ultra-Lean & Ressourcenschonend**: Nur ca. **12.8 MB RAM** und praktisch 0.0% CPU im Standby.
- 📦 **100% Single-File & Standalone**: FFmpeg Engine (mit nativer librist-Unterstützung) fest via go:embed integriert – **keine** externen DLLs oder Installationen nötig.
- 🎨 **Cyberpunk HUD Design**: Modernes Void-Dark-Theme mit Neon-Cyan & Matrix-Green Statusanzeigen.
- 🔄 **Jitter-Puffer Schnellauswahl**: Direkte Buffer-Buttons (500ms, 1000ms, 2000ms, 3000ms) für minimalen Delay oder maximale Paketwiederherstellung bei unruhigen Mobilfunk-Streams.
- 🌐 **Single-PC & 2-PC Setup**: Flexible Weiterleitung direkt an OBS auf demselben PC (127.0.0.1:10000) oder an einen separaten Streaming-PC im lokalen Netzwerk.
- 📊 **Echtzeit-Telemetrie**:
  - Live-Bitrate, FPS, empfangene Datenmenge, Paketverlust & Uptime.
  - Eingebauter Web-Server mit Live-Dashboard (/remote), OBS-Browser-Overlay (/overlay) und JSON-API (/stats).
- 📡 **NOALBS & Belabox Kompatibilität**: Standard-kompatible Stats-API mit 1-Klick-Kopierbuttons für URLs.
- 🛠️ **Autostart & Tray**:
  - Automatischer Empfangsstart beim Öffnen.
  - Windows Autostart (über Windows Registry HKCU\...\Run).
  - System Tray Minimierung.
- ⬆️ **Integrierter Update-Check**: Automatische Erkennung neuer Releases auf GitHub direkt in der App.
- 🧼 **Saubere Deinstallation**: Setup bietet automatische Deinstallation und vollständiges Entfernen aus Windows (Apps & Features).

---

## 📥 Download

Die aktuellen Versionen findest du direkt auf unserer **[GitHub Releases-Seite](https://github.com/grandstyle/rist-server/releases/latest)**:

| Datei | Typ | Beschreibung |
|---|---|---|
| **Setup-RIST-Server.exe** | **Installer (Empfohlen)** | Installiert die App, erstellt Startmenü- & Desktop-Verknüpfungen und bietet eine saubere Deinstallation. |
| **RIST-Server.exe** | **Portable** | Einfach herunterladen und per Doppelklick starten – ideal für USB-Sticks oder Testzwecke. |

---

## 🛠️ Schnellstart

### 1. Sender (Smartphone / IRL-Rucksack / Kamera)
Konfiguriere deinen RIST-Sender (z. B. Belabox, Larix Broadcaster, vMix):
- **Protokoll**: RIST
- **Ziel**: ist://<DEINE-IP>:1968 (oder dein gewählter Port)
- *(Bei Übertragung übers Internet: UDP-Port 1968 in deinem Router an den PC weiterleiten).*

### 2. RIST Server starten
- Starte RIST-Server.exe (oder starte über das Desktop-Icon nach dem Setup).
- Trage deinen gewünschten Empfangsport (Standard 1968) und die Weiterleitung ein (z. B. 127.0.0.1:10000 für lokalen Empfang).
- Wähle den gewünschten Puffer (z. B. 1000ms).
- Klicke auf **▶ STREAM EMPFANGEN**.

### 3. In OBS Studio einbinden
1. Füge in OBS Studio eine **Medienquelle (Media Source)** hinzu.
2. Deaktiviere *"Lokale Datei"*.
3. Bei **Eingabe (Input)** einfügen:
   `	ext
   udp://127.0.0.1:10000?pkt_size=1316
   `
   *(Klicke in der RIST-Server App einfach auf den **[ KOPIEREN ]**-Button neben der Ziel-Adresse, um den fertigen Link direkt in die Zwischenablage zu bekommen).*

---

## 🔗 NOALBS / Bitrate Switcher Konfiguration

Füge in deiner NOALBS config.json unter streamServers folgendes ein:

`json
"streamServers": [
  {
    "streamServer": {
      "type": "Belabox",
      "statsUrl": "http://192.168.x.x:8686/stats"
    },
    "name": "RIST-Receiver",
    "enabled": true
  }
]
`

*(Klicke auf den **[ KOPIEREN ]**-Button neben der NOALBS API Zeile in der App, um deine lokale Netzwerk-URL direkt einzufügen).*

---

## 🔨 Aus dem Quellcode bauen

Voraussetzungen:
- [Go 1.24+](https://golang.org/dl/)
- go-winres für Icons & Manifest (go install github.com/tc-hib/go-winres@latest)

`powershell
# Repository klonen
git clone https://github.com/grandstyle/rist-server.git
cd rist-server

# Vollständiger Build (Portable EXE + Setup EXE)
.\build.ps1
`

---

## 💖 Support

Wenn dir das Tool gefällt und deine Streams zuverlässiger macht, freue ich mich über einen Kaffee:

[![Ko-fi Support](https://img.shields.io/badge/Ko--fi-Spenden-ff5e5b?style=for-the-badge&logo=kofi)](https://ko-fi.com/jumpgeil)

---

## 📄 Lizenz

Dieses Projekt ist unter der **[MIT License](LICENSE)** lizenziert.
