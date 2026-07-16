package navs

import (
	"net/http"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/settings"
)

// SetAdminNavs registers the "Coffee Theme" settings entry under the admin
// Themes category. The item only appears while this plugin is the active portal
// theme, so its settings are not offered when the theme isn't in use.
func SetAdminNavs(api sdkapi.IPluginApi) {
	api.Http().Navs().AdminNavsFactory(func(r *http.Request) []sdkapi.AdminNavItemOpt {
		if !settings.IsActivePortalTheme(api) {
			return []sdkapi.AdminNavItemOpt{}
		}

		return []sdkapi.AdminNavItemOpt{
			{
				Category:  sdkapi.NavCategoryThemes,
				Label:     api.Translate("label", "Coffee Theme"),
				RouteName: "admin:coffee-theme:settings",
				Icon:      "<i class='bi bi-cup-hot'></i>",
				Order:     5000,
				Keywords: []string{
					api.Translate("label", "coffee"),
					api.Translate("label", "theme"),
					api.Translate("label", "portal"),
					api.Translate("label", "banner"),
					api.Translate("label", "logo"),
					api.Translate("label", "branding"),
					api.Translate("label", "Coffee Theme"),
				},
			},
		}
	})
}
