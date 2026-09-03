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
