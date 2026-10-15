const { test } = require('node:test');
const assert = require('node:assert/strict');
const { readFileSync } = require('node:fs');
const { runInNewContext } = require('node:vm');
const source = readFileSync(require('node:path').join(__dirname, '../assets/js/hero-carousel.js'), 'utf8');

// Exercise the real carousel with a deterministic clock and small DOM stand-ins.
function setup({ count = 3, reduced = false, failed = [], decodeDelay = 0 } = {}) {
  let time = 0, sequence = 0;
  const tasks = new Map(), images = [], listeners = {}, heroListeners = {};
  const classes = () => {
    const values = new Set();
    return { add: x => values.add(x), remove: x => values.delete(x), contains: x => values.has(x) };
  };
  const node = () => ({ classList: classes(), textContent: '', removeAttribute() {} });
  const slides = Array.from({ length: count }, (_, i) => ({ dataset: {
    photo: `photo-${i}.jpg`, name: `Animal ${i}`, country: `Country ${i}`,
    url: `/sightings/${i}`, credit: `Observer ${i}`,
  } }));
  function slide() {
    const fields = new Map(['image', 'name', 'country', 'link', 'credit', 'index'].map(x => [`[data-hero-${x}]`, node()]));
    fields.get('[data-hero-image]').src = 'photo-0.jpg';
    const s = { classList: classes(), fields, querySelector: x => fields.get(x),
      setAttribute() {}, removeAttribute() {}, cloneNode: () => slide(),
      remove() { hero.children = hero.children.filter(x => x !== s); } };
    return s;
  }
  const hero = { children: [], contains: el => el === hero,
    querySelector: x => x === '[data-hero-slides]' ? { children: slides } : hero.children[0],
    addEventListener: (type, fn) => { heroListeners[type] = fn; },
    append: x => hero.children.push(x) };
  hero.children.push(slide());
  const body = { classList: classes() };
  if (reduced) body.classList.add('no-motion');
  const document = { body, hidden: false, activeElement: null,
    querySelectorAll: () => [hero], addEventListener: (type, fn) => { listeners[type] = fn; } };
  const schedule = (fn, delay, interval = false) => {
    const id = ++sequence; tasks.set(id, { fn, at: time + delay, delay, interval }); return id;
  };
  class Image {
    set src(value) {
      images.push(value);
      if (decodeDelay) this.decode = () => new Promise(resolve => schedule(resolve, decodeDelay));
      if (failed.includes(value)) this.onerror(); else this.onload();
    }
  }
  let intersect;
  runInNewContext(source, { document, Image,
    setInterval: (fn, delay) => schedule(fn, delay, true),
    setTimeout: (fn, delay) => schedule(fn, delay),
    clearInterval: id => tasks.delete(id), clearTimeout: id => tasks.delete(id),
    queueMicrotask: fn => fn(),
    IntersectionObserver: class { constructor(fn) { intersect = fn; } observe() {} },
  });
  function tick(ms) {
    const end = time + ms;
    for (;;) {
      const next = [...tasks].filter(([, t]) => t.at <= end).sort((a, b) => a[1].at - b[1].at)[0];
      if (!next) break;
      const [id, task] = next; time = task.at;
      if (task.interval) task.at += task.delay; else tasks.delete(id);
      task.fn();
    }
    time = end;
  }
  return { hero, body, document, tick, images,
    emit: (type, event = {}) => (heroListeners[type] || listeners[type])(event),
    intersect: visible => intersect([{ isIntersecting: visible }]),
    current: () => hero.children.at(-1) };
}

test('slides every 10 seconds, updates the complete encounter, and wraps', () => {
  const x = setup();
  assert.deepEqual(x.images, ['photo-1.jpg'], 'only the next photo is preloaded');
  x.tick(9999); assert.equal(x.hero.children.length, 1);
  x.tick(1); assert.equal(x.hero.children.length, 2, 'slide starts at 10 seconds');
  const incoming = x.current();
  assert.equal(incoming.querySelector('[data-hero-image]').src, 'photo-1.jpg');
  assert.equal(incoming.querySelector('[data-hero-name]').textContent, 'Animal 1');
  assert.equal(incoming.querySelector('[data-hero-country]').textContent, 'Country 1');
  assert.equal(incoming.querySelector('[data-hero-link]').href, '/sightings/1');
  assert.equal(incoming.querySelector('[data-hero-credit]').href, '/sightings/1#photo-credits');
  assert.equal(incoming.inert, true);
  x.tick(1200); assert.equal(x.hero.children.length, 1); assert.equal(incoming.inert, false);
  x.tick(10000); assert.equal(x.current().querySelector('[data-hero-index]').textContent, '03');
  x.tick(10000); assert.equal(x.current().querySelector('[data-hero-index]').textContent, '01');
});

test('pauses for hover, keyboard focus, hidden tabs, offscreen, and reduced motion', () => {
  const x = setup();
  x.emit('pointerenter', { pointerType: 'mouse' }); x.tick(30000); assert.equal(x.hero.children.length, 1);
  x.emit('pointerleave'); x.tick(11200); assert.equal(x.current().querySelector('[data-hero-index]').textContent, '02');
  for (const reason of ['focus', 'hidden', 'offscreen', 'motion']) {
    if (reason === 'focus') { x.document.activeElement = x.hero; x.emit('focusin'); }
    if (reason === 'hidden') { x.document.hidden = true; x.emit('visibilitychange'); }
    if (reason === 'offscreen') x.intersect(false);
    if (reason === 'motion') { x.body.classList.add('no-motion'); x.emit('wildlens-motion'); }
    x.tick(30000);
    assert.equal(x.current().querySelector('[data-hero-index]').textContent, '02', reason);
    x.document.activeElement = null; x.document.hidden = false; x.intersect(true);
    x.body.classList.remove('no-motion'); x.emit('wildlens-motion');
  }
  const reduced = setup({ reduced: true }); reduced.tick(30000); assert.equal(reduced.hero.children.length, 1);
});

test('skips failed photographs and handles empty or single-animal collections', () => {
  const x = setup({ failed: ['photo-1.jpg'] });
  x.tick(11200); assert.equal(x.current().querySelector('[data-hero-name]').textContent, 'Animal 2');
  for (const count of [0, 1]) { const single = setup({ count }); single.tick(30000); assert.equal(single.hero.children.length, 1); }
});

test('pausing mid-transition settles one coherent slide without a lingering timer', () => {
  const x = setup(); x.tick(10100);
  x.body.classList.add('no-motion'); x.emit('wildlens-motion');
  assert.equal(x.hero.children.length, 1);
  assert.equal(x.current().querySelector('[data-hero-name]').textContent, 'Animal 1');
  assert.equal(x.current().inert, false);
  x.tick(30000); assert.equal(x.hero.children.length, 1);
});


test('hover lets an ongoing crossfade finish before pausing rotation', () => {
  const x = setup(); x.tick(10300);
  x.emit('pointerenter', { pointerType: 'mouse' });
  assert.equal(x.hero.children.length, 2, 'hover must not cut the fade short');
  x.tick(899); assert.equal(x.hero.children.length, 2);
  x.tick(1); assert.equal(x.hero.children.length, 1);
  assert.equal(x.current().querySelector('[data-hero-name]').textContent, 'Animal 1');
  x.tick(30000); assert.equal(x.hero.children.length, 1);
});

test('waits for the next photo to decode before showing it', async () => {
  const x = setup({ decodeDelay: 15000 });
  x.tick(10000); assert.equal(x.hero.children.length, 1, 'undecoded images stay hidden');
  x.tick(5000); await Promise.resolve();
  x.tick(5000); assert.equal(x.hero.children.length, 2);
  x.tick(1200); assert.equal(x.current().querySelector('[data-hero-name]').textContent, 'Animal 1');
});
