package handlers

import (
	"net/http"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/resources/views/portal"
)

// AdminLoginPageCtrl re-renders the login page. It backs the GET on the
// HTTPS-only /login route so a login POST that gets bounced from HTTP to HTTPS
// (302, which downgrades the method to GET) lands on the form again instead of
// a 405. Already-authenticated admins are sent straight to the dashboard.
func AdminLoginPageCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if _, err := api.Http().Auth().IsAuthenticated(r); err == nil {
			http.Redirect(w, r, "/admin", http.StatusSeeOther)
			return
		}
		csrfHTML := api.Http().Helpers().CsrfHtmlTag(r)
		forgotUrl := api.Http().Helpers().UrlForPkgRoute("com.flarego.core", "auth:send-otp")
		api.Http().Response().PortalView(w, r, sdkapi.ViewPage{
			PageContent: portal.PortalLoginPage(api, csrfHTML, sdkapi.LoginPageData{ForgotPasswordUrl: forgotUrl}),
		})
	}
}

// AdminAuthenticateCtrl authenticates an admin sign-in submitted from the
// captive portal login page. It is registered on the HTTPS-only router so
// credentials are never posted over plain HTTP.
func AdminAuthenticateCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			api.Http().Response().FlashMsg(w, r, api.Translate("error", "Invalid form data"), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")

		acct, err := api.Http().Auth().AuthenticateAdminLogin(r, username, password)
		if err != nil {
			// Do NOT html.EscapeString here -- err.Error() renders through
			// templ's own auto-escaping on the login page; pre-escaping
			// here would double-escape it. See
			// core/internal/web/controllers/auth-ctrl.go's identical fix
			// for the full reasoning.
			api.Http().Response().FlashMsg(w, r, err.Error(), sdkapi.FlashMsgError)
			http.Redirect(w, r, "/login", http.StatusSeeOther)
			return
		}

		api.Http().Auth().SignIn(w, acct)
		api.Http().Response().FlashMsg(w, r, api.Translate("info", "Logged in successfully"), sdkapi.FlashMsgSuccess)
		http.Redirect(w, r, "/admin", http.StatusSeeOther)
	}
}
