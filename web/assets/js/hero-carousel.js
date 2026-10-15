(() => {
  'use strict';

  document.querySelectorAll('[data-hero-carousel]').forEach(hero => {
    const slides = [...hero.querySelector('[data-hero-slides]').children].map(el => el.dataset);
    let current = hero.querySelector('[data-hero-current]');
    if (slides.length < 2 || !current?.querySelector('[data-hero-image]')) return;

    const INTERVAL = 10000;
    const TRANSITION = 1200;
    const unavailable = new Set();
    let index = 0, ready = null, timer = null, finishTransition = null;
    let hovered = false, visible = true;
    const canRotate = () => visible && !document.hidden && !hovered &&
      !hero.contains(document.activeElement) && !document.body.classList.contains('no-motion');

    // Preload only the next animal, so every slide starts with a ready photo.
    function prepareNext() {
      ready = null;
      let next = (index + 1) % slides.length;
      while (unavailable.has(next) && next !== index) next = (next + 1) % slides.length;
      if (next === index) return;
      const photo = new Image();
      const skip = () => { unavailable.add(next); prepareNext(); };
      photo.onload = () => {
        // Decode before fading in so a large photo never flashes in half-rendered.
        if (typeof photo.decode === 'function') {
          photo.decode().then(() => { ready = next; }).catch(skip);
        } else ready = next;
      };
      photo.onerror = skip;
      photo.src = slides[next].photo;
    }

    function advance() {
      if (!canRotate() || ready === null || finishTransition) return;
      const nextIndex = ready, data = slides[nextIndex];
      ready = null;
      const next = current.cloneNode(true);
      const image = next.querySelector('[data-hero-image]');
      image.src = data.photo;
      image.alt = `${data.name} photographed in the wild`;
      image.classList.remove('image-failed');
      image.removeAttribute('data-failed');
      next.querySelector('[data-hero-name]').textContent = data.name;
      next.querySelector('[data-hero-country]').textContent = data.country;
      next.querySelector('[data-hero-link]').href = data.url;
      const credit = next.querySelector('[data-hero-credit]');
      credit.textContent = `Photo: ${data.credit} · Credits ↗`;
      credit.href = `${data.url}#photo-credits`;
      next.querySelector('[data-hero-index]').textContent = String(nextIndex + 1).padStart(2, '0');
      next.setAttribute('aria-label', `${nextIndex + 1} of ${slides.length}`);
      next.setAttribute('aria-hidden', 'true');
      next.inert = true;
      next.classList.add('is-entering');
      current.classList.add('is-leaving');
      current.inert = true;
      hero.append(next);

      // Also settle immediately when motion is paused during a transition.
      let completion;
      finishTransition = () => {
        clearTimeout(completion);
        current.remove();
        next.classList.remove('is-entering');
        next.removeAttribute('aria-hidden');
        next.inert = false;
        current = next;
        index = nextIndex;
        finishTransition = null;
        prepareNext();
      };
      completion = setTimeout(finishTransition, TRANSITION);
    }

    function updatePlayback() {
      clearInterval(timer);
      timer = null;
      if (!canRotate()) {
        // Hover pauses the next change, while the current fade finishes smoothly.
        if (!visible || document.hidden || hero.contains(document.activeElement) ||
            document.body.classList.contains('no-motion')) finishTransition?.();
        return;
      }
      timer = setInterval(advance, INTERVAL);
    }

    hero.addEventListener('pointerenter', event => {
      if (event.pointerType === 'touch') return;
      hovered = true;
      updatePlayback();
    });
    hero.addEventListener('pointerleave', () => { hovered = false; updatePlayback(); });
    hero.addEventListener('focusin', updatePlayback);
    hero.addEventListener('focusout', () => queueMicrotask(updatePlayback));
    document.addEventListener('visibilitychange', updatePlayback);
    document.addEventListener('wildlens-motion', updatePlayback);
    new IntersectionObserver(entries => {
      visible = entries[0].isIntersecting;
      updatePlayback();
    }).observe(hero);

    prepareNext();
    updatePlayback();
  });
})();
