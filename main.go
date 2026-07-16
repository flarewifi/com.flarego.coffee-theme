//go:build !mono

// Package main is the entry point for the Coffee Shop portal theme plugin.
package main

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app"
	"com.flarego.coffee-theme/app/navs"
	"com.flarego.coffee-theme/app/theme"
)

func main() {}

func Init(api sdkapi.IPluginApi) error {
	app.SetupRoutes(api)
	navs.SetAdminNavs(api)
	theme.SetPortalTheme(api)
	return nil
}
