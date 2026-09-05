package telemetry

const htmlRemote = `<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
<meta name="apple-mobile-web-app-capable" content="yes">
<meta name="apple-mobile-web-app-status-bar-style" content="black-translucent">
<meta name="mobile-web-app-capable" content="yes">
<meta name="theme-color" content="#000000">
<link rel="manifest" href="/manifest.json">
<link rel="shortcut icon" href="/icon-192.png">
<link rel="icon" type="image/png" sizes="192x192" href="/icon-192.png">
<link rel="icon" type="image/png" sizes="512x512" href="/icon-512.png">
<link rel="apple-touch-icon" href="/icon-192.png">
<link rel="apple-touch-icon" sizes="192x192" href="/icon-192.png">
<link rel="apple-touch-icon" sizes="512x512" href="/icon-512.png">
<title>RIST Remote Control</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, Helvetica, Arial, sans-serif; -webkit-tap-highlight-color: transparent; }
html, body {
  background-color: #000000;
  color: #FFFFFF;
  width: 100%;
  min-height: 100vh;
  overflow-x: hidden;
  display: flex;
  justify-content: center;
}
.app-container {
  width: 100%;
  max-width: 420px;
  padding: 16px 16px 36px 16px;
  display: flex;
  flex-direction: column;
  gap: 14px;
}
.header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 4px 0;
}
.brand {
  font-size: 16px;
  font-weight: 900;
  letter-spacing: 1.5px;
  color: #00F0FF;
  text-shadow: 0 0 12px rgba(0, 240, 255, 0.4);
}
.status-pill {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 5px 12px;
  border-radius: 20px;
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 0.8px;
  background: rgba(38, 4, 16, 0.9);
  border: 1.5px solid #FF0055;
  color: #FF0055;
  transition: all 0.3s ease;
}
.status-pill.online {
  background: rgba(4, 38, 20, 0.9);
  border-color: #00FF80;
  color: #00FF80;
  box-shadow: 0 0 14px rgba(0, 255, 128, 0.3);
}
.status-pill.listening {
  background: rgba(40, 34, 0, 0.9);
  border-color: #FFE600;
  color: #FFE600;
  box-shadow: 0 0 14px rgba(255, 230, 0, 0.3);
}
.dot {
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: currentColor;
  box-shadow: 0 0 8px currentColor;
}
.power-btn {
  width: 100%;
  height: 64px;
  border-radius: 16px;
  font-size: 16px;
  font-weight: 800;
  letter-spacing: 1px;
  display: flex;
  align-items: center;
  justify-content: center;
  gap: 10px;
  cursor: pointer;
  border: 2px solid #00FF80;
  background: linear-gradient(180deg, #083318 0%, #03170a 100%);
  color: #FFFFFF;
  box-shadow: 0 4px 20px rgba(0, 255, 128, 0.25);
  transition: transform 0.1s, filter 0.2s;
}
.power-btn:active {
  transform: scale(0.97);
  filter: brightness(1.2);
}
.power-btn.running {
  border-color: #FF0055;
  background: linear-gradient(180deg, #3d0517 0%, #1a020a 100%);
  box-shadow: 0 4px 20px rgba(255, 0, 85, 0.3);
}
.install-btn {
  width: 100%;
  padding: 13px 16px;
  border-radius: 14px;
  font-size: 13px;
  font-weight: 800;
  letter-spacing: 0.8px;
  display: none;
  align-items: center;
  justify-content: center;
  gap: 8px;
  cursor: pointer;
  border: 1.5px solid #00F0FF;
  background: rgba(0, 240, 255, 0.08);
  color: #00F0FF;
  box-shadow: 0 0 16px rgba(0, 240, 255, 0.2);
  transition: transform 0.1s, background 0.2s;
}
.install-btn:active {
  transform: scale(0.98);
  background: rgba(0, 240, 255, 0.2);
}
.grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
}
.card {
  background: #0E121C;
  border: 1px solid #1C263A;
  border-radius: 14px;
  padding: 12px 14px;
  display: flex;
  flex-direction: column;
  justify-content: space-between;
}
.card-label {
  font-size: 10px;
  font-weight: 700;
  letter-spacing: 0.8px;
  color: #788CAF;
  text-transform: uppercase;
}
.card-val {
  font-size: 20px;
  font-weight: 900;
  color: #00F0FF;
  margin-top: 4px;
  letter-spacing: -0.5px;
}
.card-val.green {
  color: #00FF80;
  text-shadow: 0 0 12px rgba(0, 255, 128, 0.4);
}
.card-val.muted {
  color: #788CAF;
}
.card-sub {
  font-size: 10px;
  font-weight: 600;
  color: #788CAF;
  margin-top: 2px;
}
.section-title {
  font-size: 11px;
  font-weight: 800;
  letter-spacing: 1px;
  color: #00F0FF;
  text-transform: uppercase;
  margin-top: 6px;
}
.buffer-group {
  display: flex;
  gap: 6px;
  background: #0E121C;
  border: 1px solid #1C263A;
  padding: 6px;
  border-radius: 14px;
}
.buffer-pill {
  flex: 1;
  padding: 8px 0;
  text-align: center;
  font-size: 12px;
  font-weight: 700;
  color: #788CAF;
  border-radius: 10px;
  cursor: pointer;
  transition: all 0.2s;
}
.buffer-pill.active {
  background: #00F0FF;
  color: #080A0F;
  box-shadow: 0 0 12px rgba(0, 240, 255, 0.4);
}
.copy-box {
  background: #0E121C;
  border: 1px solid #1C263A;
  border-radius: 14px;
  padding: 10px 14px;
  display: flex;
  justify-content: space-between;
  align-items: center;
  cursor: pointer;
}
.copy-box:active {
  background: #141A28;
}
.copy-text {
  font-size: 11px;
  font-weight: 700;
  color: #00F0FF;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 280px;
}
.copy-badge {
  font-size: 9px;
  font-weight: 800;
  color: #00FF80;
  margin-bottom: 2px;
}
.toast {
  position: fixed;
  bottom: 24px;
  left: 50%;
  transform: translateX(-50%);
  background: #00FF80;
  color: #000;
  padding: 8px 18px;
  border-radius: 20px;
  font-size: 12px;
  font-weight: 800;
  opacity: 0;
  pointer-events: none;
  transition: opacity 0.3s;
  box-shadow: 0 0 20px rgba(0, 255, 128, 0.5);
  z-index: 1000;
}
.toast.show { opacity: 1; }
</style>
</head>
<body>
<div class="app-container">
  <div class="header">
    <div class="brand">// RIST REMOTE</div>
    <div id="status-pill" class="status-pill"><div class="dot"></div><span id="status-text">OFFLINE</span></div>
  </div>

  <button id="power-btn" class="power-btn" onclick="toggleServer()">
    <span id="power-icon">▶</span>
    <span id="power-text">STREAM EMPFANGEN</span>
  </button>

  <button id="install-btn" class="install-btn" onclick="installPWA()">
    <span>📲</span> <span>Als Web App installieren</span>
  </button>

  <div class="grid">
    <div class="card">
      <div class="card-label">Live Bitrate</div>
      <div id="val-bitrate" class="card-val muted">0.0 kbit/s</div>
      <div id="sub-bitrate" class="card-sub">Eingehend</div>
    </div>
    <div class="card">
      <div class="card-label">Framerate</div>
      <div id="val-fps" class="card-val">0.0 FPS</div>
      <div class="card-sub">Video Frame Rate</div>
    </div>
    <div class="card">
      <div class="card-label">Laufzeit</div>
      <div id="val-duration" class="card-val">00:00:00</div>
      <div class="card-sub">Aktive Sitzung</div>
    </div>
    <div class="card">
      <div class="card-label">Gesamt</div>
      <div id="val-total" class="card-val">0.00 MB</div>
      <div class="card-sub">Empfangene Daten</div>
    </div>
  </div>

  <div class="section-title">⚡ RIST Puffer (Jitter-Schutz)</div>
  <div class="buffer-group">
    <div class="buffer-pill" data-ms="500" onclick="setBuffer(500)">500ms</div>
    <div class="buffer-pill active" data-ms="1000" onclick="setBuffer(1000)">1000ms</div>
    <div class="buffer-pill" data-ms="2000" onclick="setBuffer(2000)">2000ms</div>
    <div class="buffer-pill" data-ms="3000" onclick="setBuffer(3000)">3000ms</div>
  </div>

  <div class="section-title">📋 Schnellzugriff Links</div>
  <div class="copy-box" onclick="copyLink('obs')">
    <div>
      <div class="copy-badge">🎥 OBS STREAM LINK</div>
      <div id="link-obs" class="copy-text">udp://127.0.0.1:10000?pkt_size=1316</div>
    </div>
    <span style="font-size: 16px;">📋</span>
  </div>

  <div class="copy-box" onclick="copyLink('stats')">
    <div>
      <div class="copy-badge" style="color:#FFE600;">⚡ NOALBS / STATS API</div>
      <div id="link-stats" class="copy-text">http://.../stats</div>
    </div>
    <span style="font-size: 16px;">📋</span>
  </div>
</div>

<div id="toast" class="toast">✔ Kopiert!</div>

<script>
let currentConfig = { buffer_ms: 1000, out_port: 10000 };

function vibrate() {
  if (navigator.vibrate) {
    navigator.vibrate(40);
  }
}

function showToast(text) {
  const t = document.getElementById('toast');
  t.textContent = text;
  t.classList.add('show');
  setTimeout(() => t.classList.remove('show'), 1800);
}

async function fetchStats() {
  try {
    const res = await fetch('/stats');
    if (!res.ok) return;
    const d = await res.json();

    const pill = document.getElementById('status-pill');
    const pillText = document.getElementById('status-text');
    const pBtn = document.getElementById('power-btn');
    const pIcon = document.getElementById('power-icon');
    const pText = document.getElementById('power-text');

    pill.className = 'status-pill ' + (d.status || 'offline');
    if (d.status === 'online') pillText.textContent = 'LIVE STREAMING';
    else if (d.status === 'listening') pillText.textContent = 'WARTE AUF STREAM';
    else pillText.textContent = 'OFFLINE';

    if (d.server_running) {
      pBtn.classList.add('running');
      pIcon.textContent = '⏹';
      pText.textContent = 'STREAM STOPPEN';
    } else {
      pBtn.classList.remove('running');
      pIcon.textContent = '▶';
      pText.textContent = 'STREAM EMPFANGEN';
    }

    const valBitrate = document.getElementById('val-bitrate');
    const kbps = Number(d.bitrate_kbps) || 0;
    if (d.status === 'online' && kbps > 0) {
      valBitrate.className = 'card-val green';
      if (kbps >= 1000) valBitrate.textContent = (kbps / 1000).toFixed(2) + ' Mbit/s';
      else valBitrate.textContent = kbps.toFixed(1) + ' kbit/s';
    } else {
      valBitrate.className = 'card-val muted';
      valBitrate.textContent = '0.0 kbit/s';
    }

    document.getElementById('val-fps').textContent = (Number(d.fps) || 0).toFixed(1) + ' FPS';
    document.getElementById('val-duration').textContent = d.duration || '00:00:00';
    document.getElementById('val-total').textContent = (Number(d.total_mb) || 0).toFixed(2) + ' MB';
  } catch (e) {}
}

async function fetchConfig() {
  try {
    const res = await fetch('/api/config');
    if (!res.ok) return;
    const cfg = await res.json();
    currentConfig = cfg;

    document.querySelectorAll('.buffer-pill').forEach(el => {
      const ms = Number(el.getAttribute('data-ms'));
      if (ms === cfg.buffer_ms) el.classList.add('active');
      else el.classList.remove('active');
    });

    const isLocal = !cfg.out_host || cfg.out_host === '127.0.0.1' || cfg.out_host === 'localhost';
    document.getElementById('link-obs').textContent = (isLocal ? 'udp://127.0.0.1:' : 'udp://@:') + (cfg.out_port || 10000) + '?pkt_size=1316';
    document.getElementById('link-stats').textContent = window.location.origin + '/stats';
  } catch (e) {}
}

async function toggleServer() {
  vibrate();
  try {
    const res = await fetch('/api/server/toggle', { method: 'POST' });
    const data = await res.json();
    if (data.success) {
      showToast(data.running ? '✔ Server gestartet!' : 'Server gestoppt.');
      fetchStats();
    } else if (data.error) {
      showToast('Fehler: ' + data.error);
    }
  } catch (e) {
    showToast('Verbindungsfehler');
  }
}

async function setBuffer(ms) {
  vibrate();
  currentConfig.buffer_ms = ms;
  document.querySelectorAll('.buffer-pill').forEach(el => {
    if (Number(el.getAttribute('data-ms')) === ms) el.classList.add('active');
    else el.classList.remove('active');
  });
  await fetch('/api/config', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ buffer_ms: ms })
  });
  showToast('Puffer auf ' + ms + 'ms gesetzt');
}

function copyLink(type) {
  vibrate();
  let text = '';
  if (type === 'obs') {
    text = document.getElementById('link-obs').textContent;
  } else {
    text = document.getElementById('link-stats').textContent;
  }
  navigator.clipboard.writeText(text).then(() => {
    showToast('✔ In Zwischenablage kopiert!');
  });
}

// PWA Service Worker
if ('serviceWorker' in navigator) {
  window.addEventListener('load', () => {
    navigator.serviceWorker.register('/sw.js').catch(() => {});
  });
}

// PWA Install Prompt
let deferredPrompt = null;
window.addEventListener('beforeinstallprompt', (e) => {
  e.preventDefault();
  deferredPrompt = e;
  const btn = document.getElementById('install-btn');
  if (btn) btn.style.display = 'flex';
});

async function installPWA() {
  if (!deferredPrompt) {
    showToast('Tippe im Chrome-Menü (⋮) auf "App installieren"');
    return;
  }
  vibrate();
  deferredPrompt.prompt();
  const choice = await deferredPrompt.userChoice;
  if (choice.outcome === 'accepted') {
    showToast('✔ App wird installiert!');
  }
  deferredPrompt = null;
  const btn = document.getElementById('install-btn');
  if (btn) btn.style.display = 'none';
}

window.addEventListener('appinstalled', () => {
  const btn = document.getElementById('install-btn');
  if (btn) btn.style.display = 'none';
  showToast('✔ Erfolgreich als App installiert!');
});

setInterval(fetchStats, 500);
fetchStats();
fetchConfig();
</script>
</body>
</html>`

const manifestJSON = `{
  "name": "RIST Remote Control",
  "short_name": "RIST Remote",
  "description": "Mobile Fernbedienung für den RIST Receiver Server",
  "start_url": "/remote",
  "scope": "/",
  "display": "standalone",
  "orientation": "portrait",
  "background_color": "#000000",
  "theme_color": "#000000",
  "categories": ["utilities", "video"],
  "icons": [
    {
      "src": "/icon-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icon-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "any"
    },
    {
      "src": "/icon-maskable-192.png",
      "sizes": "192x192",
      "type": "image/png",
      "purpose": "maskable"
    },
    {
      "src": "/icon-maskable-512.png",
      "sizes": "512x512",
      "type": "image/png",
      "purpose": "maskable"
    }
  ]
}`

const serviceWorkerJS = `const CACHE_NAME = 'rist-remote-v2';
const STATIC_ASSETS = [
  '/remote',
  '/manifest.json',
  '/icon-192.png',
  '/icon-512.png',
  '/icon-maskable-192.png',
  '/icon-maskable-512.png'
];

self.addEventListener('install', (event) => {
  event.waitUntil(
    caches.open(CACHE_NAME).then((cache) => cache.addAll(STATIC_ASSETS)).then(() => self.skipWaiting())
  );
});

self.addEventListener('activate', (event) => {
  event.waitUntil(
    caches.keys().then((keys) => Promise.all(
      keys.filter((k) => k !== CACHE_NAME).map((k) => caches.delete(k))
    )).then(() => self.clients.claim())
  );
});

self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  if (url.pathname.startsWith('/stats') || url.pathname.startsWith('/api/') || url.pathname.startsWith('/bitrate') || url.pathname.startsWith('/sls') || url.pathname.startsWith('/belabox')) {
    return;
  }
  event.respondWith(
    fetch(event.request).then((networkResponse) => {
      if (networkResponse && networkResponse.status === 200) {
        const resClone = networkResponse.clone();
        caches.open(CACHE_NAME).then((cache) => cache.put(event.request, resClone));
      }
      return networkResponse;
    }).catch(() => caches.match(event.request))
  );
});
`