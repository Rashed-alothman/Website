/* ============================================================
   script.js — three independent features:
     1. Typewriter effect    (reads window.ROLES set by the Go template)
     2. Scroll reveal        (Intersection Observer on .reveal elements)
     3. Nav scroll behavior  (adds .nav--scrolled class + active link tracking)
============================================================ */

'use strict';

/* ──────────────────────────────────────────────────────────────
   1. TYPEWRITER
   window.ROLES is an array set inside MainPage.html by Go:
     window.ROLES = ["Software Engineer", "Backend Developer", ...];
   This function types each role character by character, pauses,
   deletes it, then moves to the next one in a loop.
────────────────────────────────────────────────────────────── */
(function initTypewriter() {
    var roles = window.ROLES || [];
    var el    = document.getElementById('tw');

    if (!el || roles.length === 0) return;

    var roleIdx    = 0;
    var charIdx    = 0;
    var isDeleting = false;

    function tick() {
        var current = roles[roleIdx];

        // Build the string to display this tick
        el.textContent = isDeleting
            ? current.substring(0, charIdx--)
            : current.substring(0, charIdx++);

        // Decide the next delay and state transition
        var delay;

        if (!isDeleting && charIdx === current.length + 1) {
            // Finished typing — pause, then start deleting
            isDeleting = true;
            delay = 2200;
        } else if (isDeleting && charIdx < 0) {
            // Finished deleting — move to next role
            isDeleting = false;
            charIdx    = 0;
            roleIdx    = (roleIdx + 1) % roles.length;
            delay      = 500;
        } else {
            // Mid-type or mid-delete
            delay = isDeleting ? 45 : 100;
        }

        setTimeout(tick, delay);
    }

    tick();
}());


/* ──────────────────────────────────────────────────────────────
   2. SCROLL REVEAL
   Any element with class .reveal starts invisible (CSS handles
   the initial opacity:0 / translateY(28px)).
   When it enters the viewport, we add .is-visible and CSS
   transitions it into view. We unobserve after revealing so
   the animation only fires once.
────────────────────────────────────────────────────────────── */
(function initReveal() {
    var items = document.querySelectorAll('.reveal');

    if (!('IntersectionObserver' in window)) {
        // Fallback: just show everything for older browsers
        items.forEach(function(el) { el.classList.add('is-visible'); });
        return;
    }

    var observer = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
            if (entry.isIntersecting) {
                entry.target.classList.add('is-visible');
                observer.unobserve(entry.target); // reveal once, then stop watching
            }
        });
    }, {
        threshold: 0.08 // trigger when 8% of the element is visible
    });

    items.forEach(function(el) { observer.observe(el); });
}());


/* ──────────────────────────────────────────────────────────────
   3. NAV BEHAVIOR
   a) Add .nav--scrolled when the user scrolls past 40px — CSS
      applies the blur/background effect.
   b) Track which section is currently in view and add
      .is-active to the matching nav link.
────────────────────────────────────────────────────────────── */
(function initNav() {
    var nav      = document.getElementById('nav');
    var navLinks = document.querySelectorAll('.nav__link');
    var sections = document.querySelectorAll('section[id]');

    if (!nav) return;

    /* a) Blur the nav bar on scroll */
    window.addEventListener('scroll', function() {
        nav.classList.toggle('nav--scrolled', window.scrollY > 40);
    }, { passive: true });

    /* b) Highlight the active nav link based on scroll position.
          rootMargin "-40% 0px -55% 0px" means we fire when the
          section occupies the middle band of the viewport. */
    if (!('IntersectionObserver' in window)) return;

    var sectionObserver = new IntersectionObserver(function(entries) {
        entries.forEach(function(entry) {
            if (!entry.isIntersecting) return;

            // Remove active from all links
            navLinks.forEach(function(link) {
                link.classList.remove('is-active');
            });

            // Add active to the link whose href matches this section's id
            var activeLink = document.querySelector(
                '.nav__link[href="#' + entry.target.id + '"]'
            );
            if (activeLink) activeLink.classList.add('is-active');
        });
    }, {
        rootMargin: '-40% 0px -55% 0px'
    });

    sections.forEach(function(section) { sectionObserver.observe(section); });
}());
