package handlers

import (
	"errors"
	"net/http"
	"path/filepath"
	"strings"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/settings"
	adminviews "com.flarego.coffee-theme/resources/views/admin"
)

// routeSettingsIndex is the admin settings page route this controller renders
// and redirects back to.
const routeSettingsIndex = "admin:coffee-theme:settings"

// errUnsupportedImage is returned when an uploaded file is not an allowed image type.
var errUnsupportedImage = errors.New("unsupported image type")

// allowedImageExts are the accepted brand-logo / banner upload extensions.
var allowedImageExts = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".webp": true,
	".gif":  true,
	".svg":  true,
}

// ShowSettingsCtrl renders the coffee-theme settings page inside the active
// admin theme, previewing the current logo, banner, and welcome text.
func ShowSettingsCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !guardActive(api, w, r) {
			return
		}

		cfg := settings.Get(api)
		data := adminviews.CoffeeSettingsData{
			BannerText:    cfg.BannerText,
			LogoURL:       settings.LogoURL(api, cfg),
			BannerURL:     settings.BannerURL(api, cfg),
			HasCustomLogo: cfg.LogoFile != "",
			HasCustomBanner: cfg.BannerFile != "",
		}

		api.Http().Response().AdminView(w, r, sdkapi.ViewPage{
			Assets:      sdkapi.ViewAssets{CssFile: "settings.css"},
			PageContent: adminviews.CoffeeSettingsView(api, data),
		})
	}
}

// SaveSettingsCtrl persists the welcome text and any uploaded/removed images.
func SaveSettingsCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !guardActive(api, w, r) {
			return
		}

		res := api.Http().Response()

		// 8 MiB in-memory threshold; larger parts spill to temp files. The
		// storage layer enforces the real per-file size cap on write.
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Invalid form data"), sdkapi.FlashMsgError)
			res.Redirect(w, r, routeSettingsIndex)
			return
		}

		cfg := settings.Get(api)
		cfg.BannerText = strings.TrimSpace(r.FormValue("banner_text"))

		if name, err := resolveImageField(api, r, "logo", "remove_logo", "logo", cfg.LogoFile); err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Logo must be a PNG, JPG, WEBP, GIF, or SVG image"), sdkapi.FlashMsgError)
			res.Redirect(w, r, routeSettingsIndex)
			return
		} else {
			cfg.LogoFile = name
		}

		if name, err := resolveImageField(api, r, "banner", "remove_banner", "banner", cfg.BannerFile); err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Banner must be a PNG, JPG, WEBP, GIF, or SVG image"), sdkapi.FlashMsgError)
			res.Redirect(w, r, routeSettingsIndex)
			return
		} else {
			cfg.BannerFile = name
		}

		if err := settings.Save(api, &cfg); err != nil {
			api.Logger().Error("coffee-theme: failed to save portal settings: " + err.Error())
			res.FlashMsg(w, r, api.Translate("error", "Unable to save theme settings"), sdkapi.FlashMsgError)
			res.Redirect(w, r, routeSettingsIndex)
			return
		}

		res.FlashMsg(w, r, api.Translate("success", "Theme settings saved"), sdkapi.FlashMsgSuccess)
		res.Redirect(w, r, routeSettingsIndex)
	}
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

// guardActive blocks the settings page/route unless this plugin is the active
// captive-portal theme, so the settings are only reachable while the theme is
// selected (even by direct URL, not just via the hidden nav item). Returns true
// when the request may proceed.
func guardActive(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request) bool {
	if settings.IsActivePortalTheme(api) {
		return true
	}
	api.Http().Response().FlashMsg(w, r,
		api.Translate("info", "Select the Coffee Shop theme as your captive portal theme to configure it"),
		sdkapi.FlashMsgInfo)
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
	return false
}

// resolveImageField applies one image field's form input to the stored config
// and returns the resulting stored filename ("" == use bundled default):
//   - remove checkbox ticked  -> delete any custom file, return "".
//   - a file uploaded         -> validate + store it (deleting a stale
//     differently-named previous file), return the new filename.
//   - nothing submitted       -> return the current filename unchanged.
func resolveImageField(api sdkapi.IPluginApi, r *http.Request, fileField, removeField, storeBase, current string) (string, error) {
	if r.FormValue(removeField) == "on" {
		if current != "" {
			if err := api.Storage().Delete(current); err != nil {
				api.Logger().Error("coffee-theme: failed to delete " + storeBase + ": " + err.Error())
			}
		}
		return "", nil
	}

	file, header, err := r.FormFile(fileField)
	if err != nil {
		// No file chosen for this field: keep whatever is stored now.
		return current, nil
	}
	defer file.Close()

	ext := strings.ToLower(filepath.Ext(header.Filename))
	if !allowedImageExts[ext] {
		return current, errUnsupportedImage
	}

	newName := storeBase + ext
	if _, err := api.Storage().WriteReader(newName, file); err != nil {
		return current, err
	}

	// A previous upload with a different extension (e.g. logo.png -> logo.svg)
	// would otherwise be orphaned in storage.
	if current != "" && current != newName {
		if err := api.Storage().Delete(current); err != nil {
			api.Logger().Error("coffee-theme: failed to delete stale " + storeBase + ": " + err.Error())
		}
	}

	return newName, nil
}
