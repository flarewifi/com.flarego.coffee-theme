package app

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/handlers"
)

func SetupRoutes(api sdkapi.IPluginApi) {
	adminR := api.Http().Router().AdminRouter(nil)
	httpsR := api.Http().Router().HttpRouter(&sdkapi.HttpRouterOpts{HttpsOnly: true})
	portalR := api.Http().Router().HttpRouter(nil)

	// Admin login is served over HTTPS only via the plugin's own auth:login route.
	// GET re-renders the form (lands a 302-downgraded login POST instead of 405);
	// POST authenticates.
	httpsR.Get("/login", handlers.AdminLoginPageCtrl(api)).Name("auth:login-page")
	httpsR.Post("/login", handlers.AdminAuthenticateCtrl(api)).Name("auth:login")

	// Coffee-theme settings (brand logo, banner image, welcome text). Rendered
	// inside whichever admin theme is active.
	adminR.Group("/coffee-theme", func(sub sdkapi.IHttpRouterInstance) {
		sub.Get("/settings", handlers.ShowSettingsCtrl(api)).Name("admin:coffee-theme:settings")
		sub.Post("/settings", handlers.SaveSettingsCtrl(api)).Name("admin:coffee-theme:settings:save")
	})

	// Portal HTMX partials refreshed on session/internet SSE events, plus the
	// theme's own voucher-redemption endpoint.
	portalR.Group("/coffee-theme", func(sub sdkapi.IHttpRouterInstance) {
		sub.Get("/status-bar", handlers.PortalStatusBarCtrl(api)).Name("portal:coffee-theme:status-bar")
		sub.Get("/session-info", handlers.PortalSessionInfoCtrl(api)).Name("portal:coffee-theme:session-info")
		sub.Get("/internet-status", handlers.PortalInternetStatusCtrl(api)).Name("portal:coffee-theme:internet-status")
		sub.Post("/voucher", handlers.PortalVoucherCtrl(api)).Name("portal:coffee-theme:voucher")
	})
}
