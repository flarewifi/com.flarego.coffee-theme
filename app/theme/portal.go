package theme

import (
	"fmt"
	"net/http"
	"strings"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/settings"
	"com.flarego.coffee-theme/resources/views/portal"
)

func SetPortalTheme(api sdkapi.IPluginApi) {
	api.Themes().NewPortalTheme(sdkapi.PortalThemeOpts{
		JsFile:       "theme.js",
		CssFile:      "theme.css",
		CssLib:       sdkapi.CssLibBootstrap5,
		PreviewImage: "images/preview.png",
		LayoutBuilder: func(w http.ResponseWriter, r *http.Request, c sdkapi.IThemeComponents) {
			// The login page reuses this layout; only the portal index gets the
			// "coffee-index" body class so its card styling never leaks onto login.
			bodyClass := "coffee-portal"
			if strings.Contains(r.URL.Path, "/portal/index") {
				bodyClass += " coffee-index"
			}
			data := portal.PortalLayoutData{API: api, Components: c, BodyClass: bodyClass}
			if err := portal.PortalLayout(data).Render(r.Context(), w); err != nil {
				fmt.Fprintf(w, "<p>Error rendering portal layout: %s</p>", err.Error())
			}
		},
		LoginPageFactory: func(w http.ResponseWriter, r *http.Request, data sdkapi.LoginPageData) sdkapi.ViewPage {
			csrfHTML := api.Http().Helpers().CsrfHtmlTag(r)
			return sdkapi.ViewPage{
				PageContent: portal.PortalLoginPage(api, csrfHTML, data),
			}
		},
		IndexPageFactory: func(w http.ResponseWriter, r *http.Request) sdkapi.ViewPage {
			cfg := settings.Get(api)

			indexData := portal.PortalIndexData{
				Navs:       api.Http().Navs().GetPortalItems(r),
				IsOnline:   api.Machine().IsOnline(),
				LogoURL:    settings.LogoURL(api, cfg),
				BannerURL:  settings.BannerURL(api, cfg),
				BannerText: settings.BannerText(api, cfg),
			}

			clnt, err := api.Http().GetClientDevice(r)
			if err != nil {
				api.Logger().Error(fmt.Sprintf("coffee-theme portal: get client device error: %v", err))
				return sdkapi.ViewPage{PageContent: portal.PortalIndexPage(api, indexData)}
			}

			indexData.DeviceMac = clnt.MacAddr()
			indexData.DeviceIP = clnt.IpAddr()

			// IsConnectedCached, not IsConnected: the latter is a real
			// firewall call that must not sit on the render path.
			// IsConnectedCached reads the already-loaded, persisted
			// client_devices.status column instead -- see its doc comment on
			// sdkapi.IClientSession.
			indexData.IsSessionRunning = clnt.Session().IsConnectedCached()
			indexData.Client = clnt
			indexData.Session = clnt.Session().SessionData()

			if indexData.IsSessionRunning {
				// Never fail the whole page over a reporting query -- fall back to
				// the active session's own totals (what SessionSummary would report
				// with zero queued behind it) on error.
				summary, err := clnt.SessionSummary(r.Context())
				if err != nil {
					session := indexData.Session
					summary = sdkapi.SessionSummary{
						IsConnected: indexData.IsSessionRunning,
						IsPaused:    session.IsPaused,
						UpdatedAt:   session.UpdatedAt,
					}
					if session.EnforcesTime() {
						summary.TimeSecs = session.TimeSecs
					}
					if session.EnforcesData() {
						summary.DataMb = session.DataMb
					}
				}
				indexData.Summary = summary
			}

			return sdkapi.ViewPage{PageContent: portal.PortalIndexPage(api, indexData)}
		},
	})
}
