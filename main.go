package main

import (
	sdkapi "sdk/api"

	"com.flarego.coffee-theme/plugin"
)

func main() {}

// Init is the symbol the .so loader resolves by name (see PluginApi.Load). It
// exists only to re-export plugin.Init from package main, which
// -buildmode=plugin requires and which nothing can import. Keep it a pure
// delegation; real work belongs in plugin/.
func Init(api sdkapi.IPluginApi) error {
	return plugin.Init(api)
}
