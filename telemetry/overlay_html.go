package telemetry

const htmlOverlay = `<!DOCTYPE html>
<html lang="de">
<head>
<meta charset="UTF-8">
<title>RIST OBS Overlay</title>
<style>
* { box-sizing: border-box; margin: 0; padding: 0; font-family: 'Segoe UI', system-ui, -apple-system, sans-serif; }
body { background: transparent; overflow: hidden; padding: 12px; }
.card {
  width: 360px;
  background: rgba(14, 18, 28, 0.88);
  border: 1px solid rgba(0, 240, 255, 0.45);
  border-radius: 12px;
  padding: 12px 16px;
  box-shadow: 0 0 24px rgba(0, 240, 255, 0.18), inset 0 0 16px rgba(0, 240, 255, 0.05);
  backdrop-filter: blur(12px);
  color: #fff;
}
.header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 8px; }
.title { font-size: 11px; font-weight: 800; letter-spacing: 1.5px; color: #00F0FF; text-transform: uppercase; }
.badge {
  display: flex; align-items: center; gap: 6px;
  padding: 3px 8px; border-radius: 12px; font-size: 10px; font-weight: 700;
  background: rgba(38, 4, 16, 0.8); border: 1px solid #FF0055; color: #FF0055;
  transition: all 0.3s;
}
.badge.online { background: rgba(4, 38, 20, 0.8); border-color: #00FF80; color: #00FF80; }
.badge.listening { background: rgba(40, 34, 0, 0.8); border-color: #FFE600; color: #FFE600; }
.dot { width: 6px; height: 6px; border-radius: 50%; background: currentColor; box-shadow: 0 0 8px currentColor; }
.main-row { display: flex; justify-content: space-between; align-items: baseline; margin-bottom: 8px; }
.bitrate-box { display: flex; align-items: baseline; gap: 6px; }
.bitrate-num { font-size: 32px; font-weight: 900; line-height: 1; color: #788CAF; letter-spacing: -0.5px; transition: color 0.3s; }
.bitrate-num.active { color: #00FF80; text-shadow: 0 0 16px rgba(0, 255, 128, 0.4); }
.bitrate-unit { font-size: 13px; font-weight: 700; color: #788CAF; text-transform: uppercase; }
.fps-box { font-size: 13px; font-weight: 700; color: #788CAF; }
.meter-bg { width: 100%; height: 5px; background: rgba(28, 38, 58, 0.8); border-radius: 3px; overflow: hidden; margin-bottom: 10px; }
.meter-fill { height: 100%; width: 0%; background: linear-gradient(90deg, #00F0FF, #00FF80); transition: width 0.4s ease; box-shadow: 0 0 10px #00FF80; }
.footer { display: flex; justify-content: space-between; font-size: 10px; font-weight: 600; color: #788CAF; }
.footer span b { color: #fff; font-weight: 700; }
</style>
</head>
<body>
<div class="card">
  <div class="header">
    <div class="title">// RIST LIVE STREAM</div>
    <div id="badge" class="badge"><div class="dot"></div><span id="badge-text">OFFLINE</span></div>
  </div>
  <div class="main-row">
    <div class="bitrate-box">
      <span id="bitrate" class="bitrate-num">0.0</span>
      <span id="unit" class="bitrate-unit">kbit/s</span>
    </div>
    <div id="fps" class="fps-box">0.0 FPS</div>
  </div>
  <div class="meter-bg"><div id="meter" class="meter-fill"></div></div>
  <div class="footer">
    <span>GESAMT: <b id="total">0.00 MB</b></span>
    <span>LAUFZEIT: <b id="duration">00:00:00</b></span>
  </div>
</div>
<script>
async function update() {
  try {
    const res = await fetch('/stats');
    if (!res.ok) return;
    const data = await res.json();
    const badge = document.getElementById('badge');
    const badgeText = document.getElementById('badge-text');
    const bitrateEl = document.getElementById('bitrate');
    const unitEl = document.getElementById('unit');
    const fpsEl = document.getElementById('fps');
    const meterEl = document.getElementById('meter');
    const totalEl = document.getElementById('total');
    const durationEl = document.getElementById('duration');

    badge.className = 'badge ' + (data.status || 'offline');
    if (data.status === 'online') badgeText.textContent = 'LIVE';
    else if (data.status === 'listening') badgeText.textContent = 'WAITING';
    else badgeText.textContent = 'OFFLINE';

    const kbps = Number(data.bitrate_kbps) || 0;
    if (kbps >= 1000) {
      bitrateEl.textContent = (kbps / 1000).toFixed(2);
      unitEl.textContent = 'Mbit/s';
    } else {
      bitrateEl.textContent = kbps.toFixed(1);
      unitEl.textContent = 'kbit/s';
    }

    if (data.status === 'online' && kbps > 0) {
      bitrateEl.classList.add('active');
      const pct = Math.min(100, Math.max(0, (kbps / 8000) * 100));
      meterEl.style.width = pct + '%';
    } else {
      bitrateEl.classList.remove('active');
      meterEl.style.width = '0%';
    }

    fpsEl.textContent = (Number(data.fps) || 0).toFixed(1) + ' FPS';
    totalEl.textContent = (Number(data.total_mb) || 0).toFixed(2) + ' MB';
    durationEl.textContent = data.duration || '00:00:00';
  } catch (e) {}
}
setInterval(update, 400);
update();
</script>
</body>
</html>`