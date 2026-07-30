'use strict';

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
