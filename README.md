# ⚡ RIST Receiver Server (Cyberpunk Edition)

[![Release](https://img.shields.io/github/v/release/grandstyle/rist-server?style=for-the-badge&color=00f0ff&label=RELEASE)](https://github.com/grandstyle/rist-server/releases/latest)
[![Download Setup](https://img.shields.io/badge/DOWNLOAD-Setup--RIST--Server.exe-00ff80?style=for-the-badge&logo=windows)](https://github.com/grandstyle/rist-server/releases/latest/download/Setup-RIST-Server.exe)
[![Download Portable](https://img.shields.io/badge/DOWNLOAD-RIST--Server.exe%20(Portable)-00f0ff?style=for-the-badge&logo=windows)](https://github.com/grandstyle/rist-server/releases/latest/download/RIST-Server.exe)
[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=for-the-badge&logo=go)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)](LICENSE)
[![Support Ko-fi](https://img.shields.io/badge/Support-Ko--fi-ff5e5b?style=for-the-badge&logo=kofi)](https://ko-fi.com/jumpgeil)

Eine ultra-schlanke, portable **Single-File Windows-Anwendung (`RIST-Server.exe`)** in Go zum zuverlässigen Empfangen von **RIST (Reliable Internet Stream Protocol)** Video-Streams (z. B. von IRL-Rucksäcken, Belabox, Larix Broadcaster, Drones oder Außenkameras) und latenzarmen Weiterleiten per **UDP (MPEG-TS)** an **OBS Studio, vMix oder VLC**.

Inklusive integriertem **HTTP-Telemetrie-Server für automatische Bitrate- & Szenen-Switcher (NOALBS, Belabox, SRT Live Server / SLS, OBS-Scripts)**, einem **Live-Browser-Overlay für OBS** sowie einer **Smartphone Remote-Control Web-App (PWA)**.

---

## 📥 Direkt-Download & Installation

Lade die aktuelle Version direkt von den **[GitHub Releases](https://github.com/grandstyle/rist-server/releases/latest)** herunter:

| Datei | Direktlink | Beschreibung |
| :--- | :--- | :--- |
| 🛠️ **`Setup-RIST-Server.exe`** | **[⬇️ Installer herunterladen](https://github.com/grandstyle/rist-server/releases/latest/download/Setup-RIST-Server.exe)** | **Empfohlen:** Eigenständiges Windows-Setup mit Desktop- & Startmenü-Verknüpfung, optionalem Windows-Autostart und sauberer Deinstallation über Windows *Apps & Features*. |
| 📦 **`RIST-Server.exe`** | **[⬇️ Portable EXE herunterladen](https://github.com/grandstyle/rist-server/releases/latest/download/RIST-Server.exe)** | **100% Portable Single-File:** Keine Installation nötig – einfach herunterladen und starten. FFmpeg mit `librist` ist bereits fest in die `.exe` eingebettet. |

> [!NOTE]
> **Windows SmartScreen Hinweis:** Da die Anwendung ohne kostenpflichtiges Microsoft-EV-Zertifikat signiert ist, zeigt Windows beim ersten Start evtl. *"Der Computer wurde durch Windows geschützt"* an. Klicke dort einfach auf **„Weitere Informationen“** und anschließend auf **„Trotzdem ausführen“**.

---

## 🚀 Highlights & Features

- ⚡ **Ultra-Lean & Ressourcenschonend**:
  - Reine Win32-GDI Benutzeroberfläche ohne Browser-/Electron-Ballast.
  - Nur ca. **12.8 MB RAM** und **0.0 % CPU-Last** im Leerlauf.
- 📦 **100% Single-File & Standalone**:
  - Die komplette FFmpeg-Engine (inklusive `librist`-Protokollunterstützung) ist per `go:embed` direkt in die ausführbare Datei integriert und wird bei Bedarf automatisch verwaltet.
  - **Windows Job Objects** garantieren, dass beim Schließen der App niemals verwaiste `ffmpeg.exe`-Hintergrundprozesse übrig bleiben.
- 🎨 **Cyberpunk HUD-Design**:
  - Kontrastreiches Dark-Mode Interface mit Neon-Cyan & Matrix-Green Status-Indikatoren (`OFFLINE`, `WAITING FOR STREAM`, `LIVE STREAMING`).
- 🔄 **Jitter-Puffer Schnellwahl**:
  - Direkt umschaltbare Puffergrößen (**`500 ms`**, **`1000 ms`**, **`2000 ms`**, **`3000 ms`**) – wähle minimale Latenz für stabiles WLAN/LAN oder höheren Recovery-Puffer für wackelige Mobilfunk-/IRL-Verbindungen.
- 🌐 **Single-PC, 2-PC & LAN-Broadcast Support**:
  - **Single-PC:** Standard-Weiterleitung an `127.0.0.1:10000` (lokales OBS Studio).
  - **2-PC Setup:** Weiterleitung an jede beliebige IP im Heimnetzwerk (z. B. `192.168.1.50:10000`).
  - **Broadcast:** Unterstützt UDP-Broadcast-Adressen (`255.255.255.255` bzw. `*.255`).
- 📱 **Integrierte Smartphone Remote-Control App (PWA)**:
  - Über `http://<DEINE-LAN-IP>:8686/remote` (oder per 1-Klick-Kopierbutton) kannst du den Server bequem vom Smartphone aus starten, stoppen, Ports/Puffer ändern und Live-Bitraten überwachen.
  - Als **Progressive Web App (PWA)** direkt auf dem Android-/iOS-Homescreen installierbar.
- 🎥 **Integriertes OBS Live-Telemetrie Overlay**:
  - Fertiges Cyberpunk-HUD Browser-Source-Overlay unter `http://127.0.0.1:8686/overlay` mit Live-Bitrate (`kbit/s` / `Mbit/s`), FPS, Balkenanzeige, Datenvolumen und Stream-Laufzeit.
- 📡 **Universelle Bitrate-Switcher API (NOALBS / Belabox / SLS)**:
  - Mehrere Endpunkte (`/stats`, `/api/stats`, `/belabox`, `/sls`, `/bitrate`) für nahtlose Integration in **NOALBS** und eigene OBS-Skripte.
  - Automatische Null-Setzung der Bitrate bei Stream-Abbruch (Watchdog) für zuverlässiges Szenen-Umschalten (z. B. auf *„Verbindung verloren / BRB“*).
- 🔔 **Komfort- & Automatisierungs-Funktionen**:
  - **Audio-Warnton**: Akustisches Feedback, sobald ein Stream verbindet oder abreißt.
  - **System-Tray Integration**: Minimiert auf Wunsch sauber in den Windows-Infobereich (inkl. Rechtsklick-Menü).
  - **Doppelter Autostart**: Optionaler **Stream-Autostart** beim Öffnen des Programms + **Windows-Autostart** beim Hochfahren des PCs.

---

## 📖 Schritt-für-Schritt Anleitung

### Schritt 1: RIST-Server starten & konfigurieren
1. Starte **`RIST-Server.exe`** (oder installiere das Programm über **`Setup-RIST-Server.exe`**).
2. Im oberen Bereich **STREAM PORTS & OPTIONEN** siehst du deine aktuelle **RIST Empfangs-Adresse** (Standard: `<Deine-LAN-IP>:1968`).
3. Bei **Weiterleitung (Ziel-IP:Port)**:
   - Wenn OBS Studio auf **demselben PC** läuft, lasse den Standardwert **`127.0.0.1:10000`** (oder einfach `10000`) stehen.
   - Wenn OBS Studio auf einem **zweiten Streaming-PC** läuft, trage dessen IP-Adresse und Port ein (z. B. **`192.168.178.50:10000`**).
4. Wähle den gewünschten **RIST Puffer (Jitter/Latenz)**:
   - **`500ms`**: Sehr geringe Latenz (ideal für Heimnetzwerk / stabiles 5G).
   - **`1000ms`** *(Standard)*: Optimale Balance zwischen Latenz und Paketwiederherstellung.
   - **`2000ms` / `3000ms`**: Maximale Stabilität bei schwierigem Mobilfunkempfang (IRL unterwegs).
5. Klicke auf **`▶ STREAM EMPFANGEN`**. Der Status wechselt auf **`WAITING FOR STREAM`** (Gelb).

---

### Schritt 2: Sender (Smartphone / Kamera / Encoder) einrichten

Konfiguriere deine Streaming-App (z. B. **Larix Broadcaster**, **Belabox**, **IRL Pro**, **OBS**, **vMix**):

- **Protokoll:** `RIST` (Main / Simple Profile)
- **Ziel-URL (im selben WLAN/LAN):**
  ```text
  rist://192.168.x.x:1968
  ```
- **Ziel-URL (von unterwegs übers Internet / Mobilfunk):**
  ```text
  rist://<DEINE-ÖFFENTLICHE-IP-ODER-DYNDNS>:1968
  ```

> [!IMPORTANT]
> **Portfreigabe im Router (nur für Streams von außerhalb deines Heimnetzwerks):**
> Wenn du von unterwegs über das Handynetz nach Hause streamst, leite in deinem Router (z. B. FRITZ!Box) den **UDP-Port `1968`** an die lokale IP-Adresse deines PCs weiter.

Sobald der Sender startet, wechselt der Status im RIST Receiver Server automatisch auf **`LIVE STREAMING`** (Grün) und zeigt Bitrate, FPS, Datenmenge und Laufzeit in Echtzeit an.

---

### Schritt 3: Stream in OBS Studio einbinden

1. Öffne **OBS Studio** und klicke bei **Quellen** auf **`+`** ➔ **Medienquelle (Media Source)**.
2. Vergib einen Namen (z. B. `RIST Kamera`) und klicke auf **OK**.
3. **Deaktiviere** das Häkchen bei **`Lokale Datei` (Local File)**.
4. Klicke im RIST Receiver Server neben **OBS Media Source** auf **`[ KOPIEREN ]`** und füge die Adresse in OBS bei **Eingabe (Input)** ein:
   ```text
   udp://127.0.0.1:10000?pkt_size=1316
   ```
5. **Empfohlene OBS-Einstellungen für die Medienquelle:**
   - **Netzwerk-Pufferung (Network Buffering):** `0 MB` – `2 MB`
   - **Eingabeformat (Input Format):** `mpegts` (oder leer lassen für automatische Erkennung)
   - **Wiedergabe neu starten, wenn Quelle aktiv wird:** Aktiviert
   - **Nichts anzeigen, wenn Wiedergabe endet:** Nach Wunsch aktivieren
6. Bestätige mit **OK** – das Live-Bild erscheint sofort in OBS!

---

## 📊 NOALBS & Automatischer Bitrate- / Szenen-Switcher

Der RIST Receiver Server besitzt einen eingebauten HTTP-Telemetrie-Server (Standard-Port **`8686`**, weicht bei Belegung automatisch auf `8687–8695` aus). Damit kann **[NOALBS (Nginx OBS Automatic Low Bitrate Switching)](https://github.com/715209/nginx-obs-automatic-low-bitrate-switching)** automatisch deine OBS-Szene auf *„Low Bitrate“* oder *„Offline / Verbindungsabbruch“* umschalten, wenn dein mobiler Empfang schwankt.

### Konfiguration in NOALBS (`config.json`)

Klicke in der App neben **NOALBS / API** auf **`[ KOPIEREN ]`** und trage die URL in deiner NOALBS `config.json` ein:

#### Option A: Als `Belabox` Server-Typ (Empfohlen)
```json
"streamServers": [
  {
    "streamServer": {
      "type": "Belabox",
      "statsUrl": "http://127.0.0.1:8686/stats"
    },
    "name": "RIST-Receiver",
    "enabled": true
  }
]
```

#### Option B: Als `SRT Live Server (SLS)` Typ
```json
"streamServers": [
  {
    "streamServer": {
      "type": "SrtLiveServer",
      "statsUrl": "http://127.0.0.1:8686/sls",
      "publisher": "publishers"
    },
    "name": "RIST-Receiver",
    "enabled": true
  }
]
```

---

## 📱 Smartphone Remote-Control (PWA) & OBS Overlay

Über die unteren Kopier-Buttons im Programm erhältst du direkten Zugriff auf zwei integrierte Web-Tools:

### 1. Smartphone Fernbedienung (`/remote`)
- **URL:** `http://<DEINE-LAN-IP>:8686/remote` (oder einfach `http://<DEINE-LAN-IP>:8686/`)
- **Funktionen:**
  - Server per Touch **Starten / Stoppen**.
  - Live-Anzeige von **Bitrate (`kbit/s` / `Mbit/s`)**, **FPS**, **Status**, **Empfangenen MB** und **Laufzeit**.
  - **Einstellungen ändern:** RIST Empfangsport, Weiterleitungs-Ziel (`IP:Port`), Jitter-Puffer (`500`–`3000 ms`) sowie die Schalter für *Audio-Warnton*, *In Tray minimieren*, *Stream-Autostart* und *Windows-Autostart*.
- **Als App installieren (PWA):** Öffne den Link im Browser deines Smartphones (Chrome auf Android / Safari auf iOS) und wähle **„Zum Startbildschirm hinzufügen“ (Add to Home Screen)** – danach startet die Fernbedienung im Vollbild wie eine native App.

### 2. OBS Live-Telemetrie Overlay (`/overlay`)
- **URL:** `http://127.0.0.1:8686/overlay`
- **Einbindung in OBS Studio:**
  1. Neue Quelle **`Browser`** hinzufügen.
  2. URL: `http://127.0.0.1:8686/overlay`
  3. Breite: `390` | Höhe: `170`
  4. Zeigt ein transparentes Cyberpunk-HUD mit Live-Bitrate, FPS, Auslastungsbalken, Gesamt-MB und Stream-Dauer direkt im Stream oder auf deinem Kontroll-Monitor.

---

## 🌐 HTTP-API Endpunkte (Referenz)

Alle Endpunkte liefern `Access-Control-Allow-Origin: *` (CORS-kompatibel) und lauschen auf Port `8686`:

| Methode | Endpunkt | Format | Beschreibung |
| :--- | :--- | :--- | :--- |
| `GET` | `/stats` oder `/api/stats` | `application/json` | Vollständige Telemetrie (`server_running`, `status`, `bitrate`, `bitrate_kbps`, `bitrate_bps`, `total_bytes`, `total_mb`, `fps`, `speed`, `duration`) |
| `GET` | `/belabox` | `application/json` | Belabox-kompatibles JSON (`bitrate`, `total_bytes`, `status`) |
| `GET` | `/sls` | `application/json` | SRT Live Server kompatibles JSON (`publishers[0].bitrate`, `bytes`, `status`) |
| `GET` | `/bitrate` | `text/plain` | Aktuelle Bitrate als reine Zahl in `kbit/s` (z. B. `4520` oder `0`) |
| `GET` | `/overlay` | `text/html` | Transparentes Live-Telemetrie-Widget für OBS Browser-Quellen |
| `GET` | `/remote` | `text/html` | Mobile Remote-Control Web-App (PWA) |
| `POST` | `/api/server/toggle` | `application/json` | Schaltet den RIST-Empfang um (Start/Stop) |
| `POST` | `/api/server/start` | `application/json` | Startet den RIST-Empfang |
| `POST` | `/api/server/stop` | `application/json` | Stoppt den RIST-Empfang |
| `GET` / `POST` | `/api/config` | `application/json` | Liest oder speichert die Server-Konfiguration (startet bei laufendem Stream mit neuen Parametern neu) |

**Beispiel-Antwort von `GET /stats`:**
```json
{
  "server_running": true,
  "status": "online",
  "bitrate": 4850,
  "bitrate_kbps": 4850.4,
  "bitrate_bps": 4850000,
  "total_bytes": 154828800,
  "total_mb": 147.65625,
  "fps": 60,
  "speed": "1.00x",
  "duration": "00:04:12"
}
```

---

## ⚙️ Speicherung der Einstellungen

Alle Einstellungen werden automatisch benutzerbezogen gespeichert unter:
```text
%APPDATA%\RIST-Server\config.json
```
*(z. B. `C:\Users\<DeinName>\AppData\Roaming\RIST-Server\config.json`)*

Der optionale **Windows-Autostart** wird sauber ohne Administratorrechte in der Windows-Registry unter `HKCU\Software\Microsoft\Windows\CurrentVersion\Run` (`RISTReceiverServer`) hinterlegt.

---

## 🔨 Aus dem Quellcode kompilieren (Build from Source)

### Voraussetzungen
- **Windows 10 / 11 (64-Bit)**
- **[Go 1.24+](https://go.dev/dl/)**
- **`go-winres`** für Windows-Icons & Anwendungs-Manifest:
  ```powershell
  go install github.com/tc-hib/go-winres@latest
  ```

### Kompilieren (`RIST-Server.exe` & `Setup-RIST-Server.exe`)
```powershell
# 1. Repository klonen
git clone https://github.com/grandstyle/rist-server.git
cd rist-server

# 2. Build-Skript ausführen
# (Lädt assets\ffmpeg.exe automatisch aus den GitHub Releases herunter, falls noch nicht vorhanden)
powershell -ExecutionPolicy Bypass -File .\build.ps1
```

Nach erfolgreichem Durchlauf liegen im Hauptverzeichnis:
- **`RIST-Server.exe`** *(Portable Single-File Anwendung)*
- **`Setup-RIST-Server.exe`** *(Standalone Windows-Installer inkl. Uninstaller)*

### lokalen Test-Stream senden
Mit dem beiliegenden Skript `test_stream.ps1` kannst du jederzeit ohne externes Gerät einen lokalen Test-Videostream (SMPTE-Testbild + 1 kHz Sinuston) an `rist://127.0.0.1:1968` senden:
```powershell
powershell -ExecutionPolicy Bypass -File .\test_stream.ps1
```

---

## 📁 Projektstruktur

```text
rist-server/
├── assets/             # Eingebettete Ressourcen (go:embed: FFmpeg Engine, App- & PWA-Icons)
├── engine/             # FFmpeg/librist Prozess-Steuerung & Windows Job Objects
├── installer/          # Eigenständiger Win32-GUI Installer & Uninstaller (Setup-RIST-Server.exe)
├── settings/           # Persistente JSON-Konfiguration (%APPDATA%) & Windows-Registry Autostart
├── telemetry/          # Echtzeit-Parser, Watchdog, Audio-Alerts, HTTP-API, PWA Remote & OBS Overlay
├── ui/                 # Native Windows GDI/Win32 Cyberpunk Benutzeroberfläche & System-Tray
├── winres/             # Windows-Manifest & Datei-Metadaten für go-winres
├── build.ps1           # Vollautomatisches Build-Skript für Server & Setup
├── main.go             # Einstiegspunkt der Anwendung
└── test_stream.ps1     # Lokaler RIST-Teststream-Generator
```

---

## 💖 Support

Wenn dir der **RIST Receiver Server** gefällt und deine IRL- oder Gaming-Streams zuverlässiger macht, freue ich mich über deine Unterstützung:

[![Ko-fi Support](https://img.shields.io/badge/Ko--fi-jumpgeil%20unterst%C3%BCtzen-ff5e5b?style=for-the-badge&logo=kofi)](https://ko-fi.com/jumpgeil)

---

## 📄 Lizenz

Veröffentlicht unter der **[MIT Lizenz](LICENSE)**.
