package app

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/handlers"
)

func SetupRoutes(api sdkapi.IPluginApi) {
	adminR := api.Http().Router().AdminRouter(nil)
	portalR := api.Http().Router().HttpRouter(nil)

	// Admin login is served on the plugin's plain router, not HttpsOnly:true --
	// RootRouter is shared by both the plain-HTTP and HTTPS listeners (see
	// middlewares.RequireHTTPS's own comment), so this already reaches HTTPS
	// too whenever the page that rendered the form was itself served over
	// HTTPS. Whether the login page/form is forced onto HTTPS at all is
	// decided once, globally, by middlewares.ForceHTTPS
	// (AppConfig.ForceAdminHttps) -- this route must not impose a second,
	// unconditional HTTPS requirement on top of that, or turning
	// ForceAdminHttps off wouldn't actually let an operator sign in over
	// plain HTTP.
	// GET re-renders the login form (e.g. if this URL is ever visited
	// directly); POST authenticates.
	portalR.Get("/login", handlers.AdminLoginPageCtrl(api)).Name("auth:login-page")
	portalR.Post("/login", handlers.AdminAuthenticateCtrl(api)).Name("auth:login")

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
