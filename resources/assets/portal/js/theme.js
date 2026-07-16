'use strict';

// Live-ticks the "Time Left" value between the periodic HTMX refreshes so the
// countdown looks continuous. Re-initialised after every htmx settle because
// the session panel is swapped in by htmx.
var _timerHandle = null;

function initSessionTimer() {
  if (_timerHandle) {
    clearInterval(_timerHandle);
    _timerHandle = null;
  }

  var el = document.getElementById('session-time');
  if (!el) return;

  var secs = parseInt(el.getAttribute('data-value'), 10);
  if (isNaN(secs) || secs <= 0) return;

  _timerHandle = setInterval(function () {
    secs--;
    if (secs <= 0) {
      clearInterval(_timerHandle);
      _timerHandle = null;
      secs = 0;
    }
    el.textContent = formatTime(secs);
  }, 1000);
}

function formatTime(totalSecs) {
  var d = Math.floor(totalSecs / 86400);
  var rem = totalSecs % 86400;
  var h = Math.floor(rem / 3600);
  rem = rem % 3600;
  var m = Math.floor(rem / 60);
  var s = rem % 60;

  var parts = '';
  var started = false;
  if (d > 0) { parts += d + 'd '; started = true; }
  if (h > 0 || started) { parts += h + 'h '; started = true; }
  if (m > 0 || started) { parts += m + 'm '; }
  parts += s + 's';
  return parts;
}

document.addEventListener('DOMContentLoaded', initSessionTimer);
document.addEventListener('htmx:afterSettle', initSessionTimer);

// Ensure Enter submits the admin login form even inside the portal layout.
jQuery(document).on('keydown', '#login-form input', function (e) {
  if (e.which !== 13 && e.key !== 'Enter') return;
  e.preventDefault();
  var form = this.form;
  if (!form) return;
  if (typeof form.requestSubmit === 'function') {
    form.requestSubmit();
  } else {
    var btn = form.querySelector('button[type="submit"], [type="submit"]');
    if (btn) { btn.click(); } else { form.submit(); }
  }
});
