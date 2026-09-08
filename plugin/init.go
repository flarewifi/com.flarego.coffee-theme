package plugin

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app"
	"com.flarego.coffee-theme/app/theme"
)

func Init(api sdkapi.IPluginApi) error {
	app.SetupRoutes(api)
	theme.SetPortalTheme(api)
	return nil
}
