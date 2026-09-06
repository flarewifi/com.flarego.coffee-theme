package handlers

import (
	"errors"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/app/settings"
	adminviews "com.flarego.coffee-theme/resources/views/admin"
)

// routeSettingsIndex is the admin settings page route this controller
// renders and redirects back to; routeSettingsSave is the form it posts to.
// Both are resolved here rather than in the template, because each needs the
// variant being edited appended to it.
const (
	routeSettingsIndex = "admin:coffee-theme:settings"
	routeSettingsSave  = "admin:coffee-theme:settings:save"
)

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
		v, slug, ok := editedVariant(api, w, r)
		if !ok {
			return
		}

		cfg := settings.Get(v)
		data := adminviews.CoffeeSettingsData{
			VariantName:     v.Name(),
			SaveURL:         variantURL(api.Http().Helpers().UrlForRoute(routeSettingsSave), slug),
			BannerText:      cfg.BannerText,
			LogoURL:         settings.LogoURL(api, v, cfg),
			BannerURL:       settings.BannerURL(api, v, cfg),
			HasCustomLogo:   cfg.LogoFile != "",
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
		res := api.Http().Response()

		v, slug, ok := editedVariant(api, w, r)
		if !ok {
			return
		}

		// 8 MiB in-memory threshold; larger parts spill to temp files. The
		// storage layer enforces the real per-file size cap on write.
		if err := r.ParseMultipartForm(8 << 20); err != nil {
			res.FlashMsg(w, r, api.Translate("error", "Invalid form data"), sdkapi.FlashMsgError)
			redirectBack(api, w, r, slug)
			return
		}

		cfg := settings.Get(v)
		cfg.BannerText = strings.TrimSpace(r.FormValue("banner_text"))

		if name, err := resolveImageField(api, v, r, "logo", "remove_logo", "logo", cfg.LogoFile); err != nil {
			flashImageError(api, w, r, err, api.Translate("error", "Logo must be a PNG, JPG, WEBP, GIF, or SVG image"), "logo")
			redirectBack(api, w, r, slug)
			return
		} else {
			cfg.LogoFile = name
		}

		if name, err := resolveImageField(api, v, r, "banner", "remove_banner", "banner", cfg.BannerFile); err != nil {
			flashImageError(api, w, r, err, api.Translate("error", "Banner must be a PNG, JPG, WEBP, GIF, or SVG image"), "banner")
			redirectBack(api, w, r, slug)
			return
		} else {
			cfg.BannerFile = name
		}

		if err := settings.Save(v, &cfg); err != nil {
			api.Logger().Error("coffee-theme: failed to save portal settings: " + err.Error())
			res.FlashMsg(w, r, api.Translate("error", "Unable to save theme settings"), sdkapi.FlashMsgError)
			redirectBack(api, w, r, slug)
			return
		}

		res.FlashMsg(w, r, api.Translate("success", "Theme settings saved"), sdkapi.FlashMsgSuccess)
		redirectBack(api, w, r, slug)
	}
}

// =============================================================================
// HELPER FUNCTIONS (internal)
// =============================================================================

// resolveImageField applies one image field's form input to the stored config
// and returns the resulting stored filename ("" == use bundled default):
//   - remove checkbox ticked  -> delete any custom file, return "".
//   - a file uploaded         -> validate + store it (deleting a stale
//     differently-named previous file), return the new filename.
//   - nothing submitted       -> return the current filename unchanged.
func resolveImageField(api sdkapi.IPluginApi, v sdkapi.IThemesVariant, r *http.Request, fileField, removeField, storeBase, current string) (string, error) {
	if r.FormValue(removeField) == "on" {
		if current != "" {
			if err := v.Storage().Delete(current); err != nil {
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

	// Content-addressed, so the stored name changes exactly when the image
	// does. Variant images are served with a long cache lifetime, and a fixed
	// name would leave every browser that already loaded the portal showing the
	// previous picture until that cache expired.
	newName, err := v.Storage().WriteHashed(storeBase+ext, file)
	if err != nil {
		return current, err
	}

	// The stored name is content-addressed, so ANY different image lands under a
	// different name and the previous one would otherwise be orphaned in the
	// variant -- where it would go on counting against the variant tree's size
	// budget forever. Re-uploading the identical image yields the identical
	// name, so that case correctly deletes nothing.
	if current != "" && current != newName {
		if err := v.Storage().Delete(current); err != nil {
			api.Logger().Error("coffee-theme: failed to delete stale " + storeBase + ": " + err.Error())
		}
	}

	return newName, nil
}

// editedVariant resolves which saved look this request edits.
//
// The admin Captive Portal list links here with ?variant=<slug>; a plain visit has no
// parameter and edits whatever variant is currently applied. It reads the URL
// rather than the form, so it is valid before the multipart body is parsed and
// every redirect below can keep the operator on the same variant.
//
// A slug naming a variant this theme no longer has resolves to nothing rather
// than falling back to the current one. That case is reported and the operator
// sent to the unparameterised page -- still a real, plainly labelled variant --
// instead of quietly overwriting the live look under a deleted variant's name.
func editedVariant(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request) (sdkapi.IThemesVariant, string, bool) {
	slug := strings.TrimSpace(r.URL.Query().Get("variant"))
	v := api.Themes().GetVariant(slug)

	// An unresolvable variant with no slug asked for means the theme could not
	// resolve a CURRENT variant either. There is nowhere to send the operator
	// -- the redirect below lands right back here -- so fall through: the page
	// then shows the theme's own defaults and the save refuses, which is
	// visible without being a redirect loop.
	if v.Slug() != "" || slug == "" {
		return v, slug, true
	}

	res := api.Http().Response()
	res.FlashMsg(w, r, api.Translate("error", "That variant no longer exists."), sdkapi.FlashMsgError)
	redirectBack(api, w, r, "")
	return nil, "", false
}

// variantURL appends the variant being edited to one of this page's own URLs.
// An empty slug means "whichever variant is current", which is deliberately
// left unpinned: the page then follows the applied variant rather than freezing
// on the slug that happened to be current when the link was built.
func variantURL(base, slug string) string {
	if slug == "" {
		return base
	}
	return base + "?variant=" + url.QueryEscape(slug)
}

// redirectBack returns to this settings page with the variant intact.
//
// A plain 303 rather than IHttpResponse.Redirect because that helper resolves a
// route NAME and cannot carry a query string; this form is an ordinary HTML
// POST, not htmx, so there is no Hx-Redirect handling to lose.
func redirectBack(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request, slug string) {
	base := api.Http().Helpers().UrlForRoute(routeSettingsIndex)
	http.Redirect(w, r, variantURL(base, slug), http.StatusSeeOther)
}

// flashImageError reports the right one of the two ways storing an image can
// fail. Only one of them is the operator's doing.
//
// resolveImageField refuses an unsupported extension itself, but it also
// forwards whatever the storage layer returns -- the variant tree's byte cap,
// or an ordinary disk error. Reporting those as "must be a PNG" sends the
// operator off to re-encode a file that was never the problem, so they are
// logged and named for what they are.
func flashImageError(api sdkapi.IPluginApi, w http.ResponseWriter, r *http.Request, err error, typeMsg, field string) {
	res := api.Http().Response()
	if errors.Is(err, errUnsupportedImage) {
		res.FlashMsg(w, r, typeMsg, sdkapi.FlashMsgError)
		return
	}
	api.Logger().Error("coffee-theme: failed to store " + field + ": " + err.Error())
	res.FlashMsg(w, r, api.Translate("error", "Unable to save the image"), sdkapi.FlashMsgError)
}
