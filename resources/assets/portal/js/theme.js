'use strict';

// A device can have more than one session (an active one plus any number
// still queued behind it) -- see .templ's SessionInfo. Core's ticker only
// ever counts down the ACTIVE session's own value, so the .templ view bakes
// the total contributed by whatever is still queued into
// data-extra-time-secs/data-extra-data-mb, and this adds it back on top of
// core's value (its own resync/tick logic untouched) so what's displayed is
// the device's TRUE remaining total, not just the active session's share.
window.$flare.ui.liveSession.setRenderer(function (el, text, value, kind) {
  if (kind === 'time' || kind === 'data') {
    var attr = kind === 'data' ? 'data-extra-data-mb' : 'data-extra-time-secs';
    var extra = Number(el.getAttribute(attr));
    if (isFinite(extra) && extra > 0) {
      var total = value + extra;
      text = kind === 'data'
        ? window.$flare.ui.liveSession.formatByteData(total)
        : window.$flare.ui.liveSession.formatTimeSecs(total);
    }
  }
  el.textContent = text;
});

// Ensure Enter submits the admin login form even inside the portal layout.
document.addEventListener('keydown', function (e) {
  if (e.key !== 'Enter' && e.keyCode !== 13) return;
  var input = e.target && e.target.closest ? e.target.closest('#login-form input') : null;
  if (!input) return;
  e.preventDefault();
  var form = input.form;
  if (!form) return;
  if (typeof form.requestSubmit === 'function') {
    form.requestSubmit();
  } else {
    var btn = form.querySelector('button[type="submit"], [type="submit"]');
    if (btn) { btn.click(); } else { form.submit(); }
  }
});

// Device code (issue #4230). The eye button in the wifi credits box shows and
// hides the code; htmx re-swaps that box on session events, so the handlers
// are delegated on document and the state is re-applied after each settle. It
// starts hidden on every page load.
(function () {
  var deviceCodeShown = false;

  function applyDeviceCode() {
    var rows = document.querySelectorAll('[data-device-code]');
    for (var i = 0; i < rows.length; i++) {
      var row = rows[i];
      var value = row.querySelector('[data-device-code-value]');
      var btn = row.querySelector('[data-device-code-toggle]');
      if (!value || !btn) continue;
      var code = row.getAttribute('data-device-code') || '';
      var showIcon = row.querySelector('[data-device-code-icon="show"]');
      var hideIcon = row.querySelector('[data-device-code-icon="hide"]');
      value.textContent = deviceCodeShown
        ? code.slice(0, 3) + ' ' + code.slice(3)
        : '••••••';
      btn.setAttribute('aria-label',
        btn.getAttribute(deviceCodeShown ? 'data-label-hide' : 'data-label-show') || '');
      if (showIcon) showIcon.hidden = deviceCodeShown;
      if (hideIcon) hideIcon.hidden = !deviceCodeShown;
    }
  }

  document.addEventListener('click', function (e) {
    var btn = e.target && e.target.closest ? e.target.closest('[data-device-code-toggle]') : null;
    if (!btn) return;
    deviceCodeShown = !deviceCodeShown;
    applyDeviceCode();
  });

  // Keep the restore input to 6 upper-case letters and digits as it is typed
  // or pasted, so a code copied as "K7M Q4P" fits.
  document.addEventListener('input', function (e) {
    var el = e.target;
    if (!el || el.id !== 'device-code-input') return;
    var v = String(el.value || '').toUpperCase().replace(/[^A-Z0-9]/g, '').slice(0, 6);
    if (el.value !== v) el.value = v;
  });

  document.addEventListener('htmx:afterSettle', applyDeviceCode);
})();
