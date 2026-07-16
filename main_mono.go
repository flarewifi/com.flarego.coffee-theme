/**
WARNING: This file system-generated, do not edit nor commit into your repo.
Edit the main.go file instead to updated this file.
*/

//go:build mono

package com_flarego_coffee_theme

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app"
	"com.flarego.coffee-theme/app/navs"
	"com.flarego.coffee-theme/app/theme"
)

func Init(api sdkapi.IPluginApi) error {
	app.SetupRoutes(api)
	navs.SetAdminNavs(api)
	theme.SetPortalTheme(api)
	return nil
}
