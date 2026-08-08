// Package settings persists and resolves the coffee-shop theme's
// admin-editable portal content: the welcome/banner text, the brand logo, and
// the banner image.
//
// Text is stored as a JSON blob in the plugin config store
// (api.Config().Plugin()); uploaded images are stored as files in the plugin
// storage area (api.Storage()). The config only records the stored filename of
// each image (empty == fall back to the bundled default), so mutable uploads
// and immutable bundled defaults never get confused.
package settings

import (
	"encoding/json"

	sdkapi "sdk/api"
)

const configKey = "portal_settings"

// Bundled default assets, served from resources/assets/public via PublicPath.
const (
	defaultLogo   = "images/logo.svg"
	defaultBanner = "images/banner.svg"
)

// Settings is the admin-editable portal content.
type Settings struct {
	// BannerText is the welcome line shown under the banner image. Empty means
	// "use the translated default" (resolved at render time, not stored).
	BannerText string `json:"banner_text"`
	// LogoFile is the stored filename of a custom brand logo in plugin storage.
	// Empty means "use the bundled default logo".
	LogoFile string `json:"logo_file"`
	// BannerFile is the stored filename of a custom banner image in plugin
	// storage. Empty means "use the bundled default banner".
	BannerFile string `json:"banner_file"`
}

// Get returns the saved settings, or a zero-value Settings (all defaults) when
// nothing has been saved yet or the stored blob is unreadable.
func Get(api sdkapi.IPluginApi) Settings {
	var s Settings

	b, err := api.Config().Plugin().Read(configKey)
	if err != nil {
		return s
	}

	if err := json.Unmarshal(b, &s); err != nil {
		return Settings{}
	}

	return s
}

// Save persists the given settings.
func Save(api sdkapi.IPluginApi, s *Settings) error {
	b, err := json.Marshal(s)
	if err != nil {
		return err
	}

	return api.Config().Plugin().Write(configKey, b)
}

// LogoURL resolves the brand-logo URL: the uploaded file when present,
// otherwise the bundled default.
func LogoURL(api sdkapi.IPluginApi, s Settings) string {
	if s.LogoFile != "" && api.Storage().Exists(s.LogoFile) {
		return api.Storage().UrlFor(s.LogoFile)
	}
	return api.Http().Helpers().PublicPath(defaultLogo)
}

// BannerURL resolves the banner-image URL: the uploaded file when present,
// otherwise the bundled default.
func BannerURL(api sdkapi.IPluginApi, s Settings) string {
	if s.BannerFile != "" && api.Storage().Exists(s.BannerFile) {
		return api.Storage().UrlFor(s.BannerFile)
	}
	return api.Http().Helpers().PublicPath(defaultBanner)
}

// BannerText resolves the welcome text: the admin-set value when present,
// otherwise a translated default.
func BannerText(api sdkapi.IPluginApi, s Settings) string {
	if s.BannerText != "" {
		return s.BannerText
	}
	return api.Translate("label", "Grab a cup, sit back, and enjoy free WiFi")
}
