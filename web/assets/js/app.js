(() => {
  'use strict';
  const reduced = window.matchMedia('(prefers-reduced-motion: reduce)');
  let paused = false;
  try { paused = localStorage.getItem('wildlens-motion') === 'paused'; } catch (_) {}
  const motionButton = document.querySelector('.motion-toggle');
  function updateMotion() {
    const off = paused || reduced.matches;
    document.body.classList.toggle('no-motion', off);
    if (motionButton) {
      motionButton.setAttribute('aria-pressed', String(off));
      motionButton.disabled = reduced.matches;
      motionButton.setAttribute('aria-label', reduced.matches ? 'Animations paused by your system preference' : off ? 'Enable 3D animations' : 'Pause 3D animations');
      motionButton.title = motionButton.getAttribute('aria-label');
    }
    document.dispatchEvent(new Event('wildlens-motion'));
  }
  motionButton?.addEventListener('click', () => {
    paused = !document.body.classList.contains('no-motion');
    // The operating system's reduced-motion preference takes priority.
    if (reduced.matches) paused = true;
    try { localStorage.setItem('wildlens-motion', paused ? 'paused' : 'enabled'); } catch (_) {}
    updateMotion();
  });
  reduced.addEventListener('change', updateMotion);
  updateMotion();

  if (window.matchMedia('(hover: hover) and (pointer: fine)').matches) {
    document.querySelectorAll('[data-tilt]').forEach(el => {
      let frame = 0;
      el.addEventListener('pointermove', event => {
        if (document.body.classList.contains('no-motion')) return;
        cancelAnimationFrame(frame);
        frame = requestAnimationFrame(() => {
          const box = el.getBoundingClientRect();
          const strength = el.classList.contains('hero') ? 1.8 : 4;
          const x = ((event.clientX - box.left) / box.width - .5) * strength;
          const y = ((event.clientY - box.top) / box.height - .5) * strength;
          el.style.transform = `perspective(1100px) rotateX(${-y}deg) rotateY(${x}deg) translateY(-2px)`;
        });
      });
      const reset = () => { cancelAnimationFrame(frame); el.style.transform = ''; };
      el.addEventListener('pointerleave', reset);
      document.addEventListener('wildlens-motion', reset);
    });
  }

  // Recover gracefully from unavailable observation photos.
  const recoverImage = img => {
    if (img.dataset.failed) return;
    img.dataset.failed = 'true';
    img.src = '/assets/images/mark.svg';
    img.classList.add('image-failed');
    img.alt += ' — photograph unavailable';
  };
  document.querySelectorAll('img').forEach(img => {
    img.addEventListener('error', () => recoverImage(img));
    if (img.complete && img.naturalWidth === 0) recoverImage(img);
  });

  function initMap(el) {
    const records = [...(document.getElementById(el.dataset.points)?.children || [])];
    if (!window.L || records.length === 0) {
      el.querySelector('.map-fallback').textContent = records.length ? 'The map is unavailable. Browse the sighting details below.' : 'No sightings to map. Try a different filter.';
      return;
    }
    el.querySelector('.map-fallback')?.remove();
    const off = document.body.classList.contains('no-motion');
    const map = L.map(el, { scrollWheelZoom: false, worldCopyJump: true, zoomAnimation: !off, fadeAnimation: !off });
    const tiles = L.tileLayer('https://tile.openstreetmap.org/{z}/{x}/{y}.png', {
      attribution: '&copy; <a href="https://www.openstreetmap.org/copyright">OpenStreetMap</a> contributors',
      maxZoom: 18, minZoom: 1
    }).addTo(map);
    let failedTiles = 0;
    tiles.on('tileerror', () => {
      if (++failedTiles === 3 && !el.querySelector('.tile-message')) {
        const message = document.createElement('p');
        message.className = 'tile-message';
        message.textContent = 'Map tiles are unavailable. Sighting markers and record links are still accessible.';
        el.append(message);
      }
    });
    const group = L.featureGroup();
    records.forEach((record, index) => {
      const d = record.dataset, lat = Number(d.lat), lng = Number(d.lng);
      if (!Number.isFinite(lat) || !Number.isFinite(lng) || Math.abs(lat) > 90 || Math.abs(lng) > 180) return;
      const content = document.createElement('div');
      content.className = 'map-popup';
      const title = document.createElement('strong'); title.textContent = d.title;
      const place = document.createElement('span'); place.textContent = d.place;
      const date = document.createElement('span'); date.textContent = d.date;
      const link = document.createElement('a'); link.href = d.url; link.textContent = 'View sighting →';
      content.append(title, place, date, link);
      // A divIcon keeps markers keyboard accessible and avoids external image assets.
      const icon = L.divIcon({ className: 'wildlife-marker', html: '<span></span>', iconSize: [16, 16], iconAnchor: [8, 8] });
      const marker = L.marker([lat, lng], { icon, title: `${d.title} · ${d.place} · ${d.date}`, alt: `Sighting ${index + 1}: ${d.title}`, keyboard: true }).bindPopup(content).addTo(group);
      marker.on('popupopen', () => { setTimeout(() => link.focus({ preventScroll: true }), 0); });
    });
    if (group.getLayers().length) {
      group.addTo(map);
      map.fitBounds(group.getBounds(), { padding: [35, 35], maxZoom: 10, animate: false });
    } else map.setView([0, 15], 2);
    map.on('focus', () => map.scrollWheelZoom.enable());
    map.on('blur', () => map.scrollWheelZoom.disable());
    new ResizeObserver(() => map.invalidateSize({ animate: false })).observe(el);
  }
  document.querySelectorAll('[data-sightings-map]').forEach(initMap);

  // A lightweight, genuinely 3D globe drawn with an orthographic projection.
  // Points rotate around the vertical axis and disappear behind the sphere.
  document.querySelectorAll('[data-globe]').forEach(canvas => {
    const ctx = canvas.getContext('2d');
    if (!ctx) return;
    let width = 0, height = 0, angle = -.45, frame = 0, visible = true;
    // Coarse geographic silhouettes used only for the decorative globe.
    const land = [
      [[-17,36],[8,38],[33,31],[43,12],[51,10],[40,-13],[32,-33],[18,-35],[10,-15],[-2,5],[-17,15]],
      [[-10,36],[-10,60],[15,72],[42,68],[60,55],[96,75],[155,65],[178,51],[140,35],[122,20],[104,0],[80,7],[70,24],[45,30],[30,40],[12,40]],
      [[-167,65],[-130,70],[-115,80],[-62,60],[-54,48],[-83,25],[-90,15],[-106,23],[-126,50]],
      [[-82,12],[-68,10],[-48,-1],[-36,-8],[-43,-24],[-60,-55],[-75,-40],[-80,-10]],
      [[112,-11],[137,-10],[154,-25],[147,-39],[116,-34]],
      [[44,-12],[51,-16],[48,-26],[44,-24]],
      [[-53,60],[-40,60],[-20,77],[-42,84],[-63,77]]
    ];
    function inside(lng, lat, polygon) {
      let found = false;
      for (let i = 0, j = polygon.length - 1; i < polygon.length; j = i++) {
        const [xi, yi] = polygon[i], [xj, yj] = polygon[j];
        if (((yi > lat) !== (yj > lat)) && lng < (xj - xi) * (lat - yi) / (yj - yi) + xi) found = !found;
      }
      return found;
    }
    const points = [];
    for (let lat = -58; lat < 82; lat += 3) {
      const step = 3 / Math.max(.25, Math.cos(lat * Math.PI / 180));
      for (let lng = -180; lng < 180; lng += step) {
        if (land.some(poly => inside(lng, lat, poly))) points.push([lng * Math.PI / 180, lat * Math.PI / 180]);
      }
    }
    function project(lng, lat, radius) {
      const longitude = lng + angle;
      const x = Math.cos(lat) * Math.sin(longitude);
      const y = Math.sin(lat);
      const z = Math.cos(lat) * Math.cos(longitude);
      const tilt = -.12;
      return { x: width * .54 + radius * x, y: height * .51 - radius * (y * Math.cos(tilt) - z * Math.sin(tilt)), z: z * Math.cos(tilt) + y * Math.sin(tilt) };
    }
    function draw() {
      ctx.clearRect(0, 0, width, height);
      const radius = Math.min(height * .46, width * .31), cx = width * .54, cy = height * .51;
      ctx.save();
      ctx.translate(cx, height * .97); ctx.scale(1, .14);
      const shadow = ctx.createRadialGradient(0, 0, 0, 0, 0, radius);
      shadow.addColorStop(0, '#738b6030'); shadow.addColorStop(1, '#738b6000');
      ctx.fillStyle = shadow; ctx.beginPath(); ctx.arc(0, 0, radius, 0, Math.PI * 2); ctx.fill(); ctx.restore();
      const ocean = ctx.createRadialGradient(cx - radius * .35, cy - radius * .4, radius * .1, cx, cy, radius);
      ocean.addColorStop(0, '#f9faed'); ocean.addColorStop(.65, '#e5eacb'); ocean.addColorStop(1, '#c3d0a6');
      ctx.fillStyle = ocean; ctx.beginPath(); ctx.arc(cx, cy, radius, 0, Math.PI * 2); ctx.fill();
      ctx.strokeStyle = '#9ba97e30'; ctx.lineWidth = .5;
      for (let lat = -60; lat <= 60; lat += 30) {
        ctx.beginPath(); let pen = false;
        for (let lng = -180; lng <= 180; lng += 3) {
          const p = project(lng * Math.PI / 180, lat * Math.PI / 180, radius);
          if (p.z > 0) { if (!pen) ctx.moveTo(p.x, p.y); else ctx.lineTo(p.x, p.y); pen = true; } else pen = false;
        } ctx.stroke();
      }
      for (let lng = -180; lng < 180; lng += 30) {
        ctx.beginPath(); let pen = false;
        for (let lat = -90; lat <= 90; lat += 3) {
          const p = project(lng * Math.PI / 180, lat * Math.PI / 180, radius);
          if (p.z > 0) { if (!pen) ctx.moveTo(p.x, p.y); else ctx.lineTo(p.x, p.y); pen = true; } else pen = false;
        } ctx.stroke();
      }
      points.forEach(([lng, lat]) => {
        const p = project(lng, lat, radius);
        if (p.z < 0) return;
        ctx.fillStyle = `rgba(87,116,62,${.3 + p.z * .45})`;
        ctx.beginPath(); ctx.arc(p.x, p.y, .7 + p.z * .45, 0, Math.PI * 2); ctx.fill();
      });
      [[37, 0], [23, -29], [78, 25], [-65, -15]].forEach(([lng, lat]) => {
        const p = project(lng * Math.PI / 180, lat * Math.PI / 180, radius);
        if (p.z < 0) return;
        ctx.strokeStyle = '#ca886f70'; ctx.lineWidth = 1; ctx.beginPath(); ctx.arc(p.x, p.y, 5.5, 0, Math.PI * 2); ctx.stroke();
        ctx.fillStyle = '#ca886f'; ctx.beginPath(); ctx.arc(p.x, p.y, 2.3, 0, Math.PI * 2); ctx.fill();
      });
    }
    let previous = 0;
    function animate(now) {
      if (document.hidden || !visible || document.body.classList.contains('no-motion')) { frame = 0; previous = 0; return; }
      angle += Math.min(now - (previous || now), 50) * .00007;
      previous = now; draw(); frame = requestAnimationFrame(animate);
    }
    function resume() { draw(); if (!frame) frame = requestAnimationFrame(animate); }
    new ResizeObserver(() => {
      const rect = canvas.getBoundingClientRect(), ratio = Math.min(window.devicePixelRatio || 1, 2);
      width = rect.width; height = rect.height; canvas.width = width * ratio; canvas.height = height * ratio; ctx.setTransform(ratio, 0, 0, ratio, 0, 0); resume();
    }).observe(canvas);
    new IntersectionObserver(entries => { visible = entries[0].isIntersecting; if (visible) resume(); }).observe(canvas);
    document.addEventListener('visibilitychange', resume);
    document.addEventListener('wildlens-motion', resume);
  });
})();
