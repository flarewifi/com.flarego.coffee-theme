package navs

import (
	"net/http"

	sdkapi "sdk/api"
)

// SetAdminNavs registers the "Coffee Theme" settings entry under the admin
// Themes category. Always shown, regardless of whether this plugin is the
// currently active portal theme -- an operator can configure a theme's
// settings ahead of switching to it.
func SetAdminNavs(api sdkapi.IPluginApi) {
	api.Http().Navs().AdminNavsFactory(func(r *http.Request) []sdkapi.AdminNavItemOpt {
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
