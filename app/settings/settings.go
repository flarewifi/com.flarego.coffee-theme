// Package settings persists and resolves the coffee-shop theme's
// admin-editable portal content: the welcome/banner text, the brand logo, and
// the banner image.
//
// Text is a JSON blob in the theme variant store (api.Themes().Config());
// uploaded images are files in the matching variant file store
// (api.Themes().Storage()). Reading and writing through those, rather than the
// plugin-wide stores, is what gives each saved variant its own copy of both.
//
// The config records only each image's stored filename (empty == fall back to
// the bundled default), so mutable uploads and immutable bundled defaults never
// get confused.
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

	b, err := api.Themes().Config().Read(configKey)
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

	return api.Themes().Config().Write(configKey, b)
}

// LogoURL resolves the brand-logo URL: the uploaded file when present,
// otherwise the bundled default.
func LogoURL(api sdkapi.IPluginApi, s Settings) string {
	if s.LogoFile != "" && api.Themes().Storage().Exists(s.LogoFile) {
		return api.Themes().Storage().UrlFor(s.LogoFile)
	}
	return api.Http().Helpers().PublicPath(defaultLogo)
}

// BannerURL resolves the banner-image URL: the uploaded file when present,
// otherwise the bundled default.
func BannerURL(api sdkapi.IPluginApi, s Settings) string {
	if s.BannerFile != "" && api.Themes().Storage().Exists(s.BannerFile) {
		return api.Themes().Storage().UrlFor(s.BannerFile)
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

// MigrateLegacySettings copies this theme's pre-variant settings -- the config
// blob and the images it names -- from the plugin-wide stores into the
// currently-applied variant. Core no longer does this for us: only the theme
// itself knows which of its config keys describe the portal look.
//
// One-time and best-effort: every failure is logged and skipped, because this
// runs during plugin registration and must never abort startup. The legacy
// files are left in place as a safety net.
func MigrateLegacySettings(api sdkapi.IPluginApi) {
	if _, err := api.Themes().Config().Read(configKey); err == nil {
		// Already migrated -- the key's presence in the variant store IS the marker.
		return
	}

	b, err := api.Config().Plugin().Read(configKey)
	if err != nil {
		// Nothing was ever saved pre-variant, so there is nothing to migrate.
		return
	}

	var s Settings
	if err := json.Unmarshal(b, &s); err != nil {
		api.Logger().Error("coffee-theme: legacy settings are unreadable, skipping migration: " + err.Error())
		return
	}

	for _, name := range []string{s.LogoFile, s.BannerFile} {
		copyLegacyImage(api, name)
	}

	// Written LAST, deliberately: this key IS the already-migrated marker, so
	// writing it before the images would make a crash in between lose them
	// permanently.
	if err := api.Themes().Config().Write(configKey, b); err != nil {
		api.Logger().Error("coffee-theme: failed to migrate settings into the theme variant: " + err.Error())
	}
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

// copyLegacyImage copies one stored upload from plugin storage into the current
// variant's storage. A missing or unreadable file is skipped rather than fatal:
// the URL resolvers already fall back to the bundled default.
func copyLegacyImage(api sdkapi.IPluginApi, name string) {
	if name == "" || !api.Storage().Exists(name) {
		return
	}

	data, err := api.Storage().Read(name)
	if err != nil {
		api.Logger().Error("coffee-theme: failed to read legacy image " + name + ": " + err.Error())
		return
	}

	if _, err := api.Themes().Storage().Write(name, data); err != nil {
		api.Logger().Error("coffee-theme: failed to migrate image " + name + ": " + err.Error())
	}
}
