package handlers

import (
	"fmt"
	"net/http"

	sdkapi "sdk/api"

	"com.flarego.coffee-theme/resources/views/portal"
)

// PortalStatusBarCtrl re-renders the connected/disconnected status pill on
// session connect/disconnect SSE events.
func PortalStatusBarCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		isConnected := false

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			api.Logger().Error(fmt.Sprintf("coffee-theme status-bar: get client device error: %v", err))
		} else {
			isConnected = clnt.Session().IsConnected()
		}

		portal.StatusBar(api, isConnected).Render(r.Context(), w)
	}
}

// PortalSessionInfoCtrl re-renders the wifi credits box: time/data left and
// the device code.
func PortalSessionInfoCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var isRunning bool
		var session sdkapi.SessionData
		var summary sdkapi.SessionSummary

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			api.Logger().Error(fmt.Sprintf("coffee-theme session-info: get client device error: %v", err))
		} else {
			isRunning = clnt.Session().IsConnected()
			session = clnt.Session().SessionData()

			// Read for a phone that is not connected too (see the index page).
			// Never fail this partial over a reporting query -- fall back to the
			// active session's own totals on error.
			var sErr error
			summary, sErr = clnt.SessionSummary(r.Context())
			if sErr != nil {
				api.Logger().Error(fmt.Sprintf("coffee-theme session-info: session summary error: %v", sErr))
				summary = sdkapi.SessionSummary{
					IsConnected: isRunning,
					IsPaused:    session.IsPaused,
					UpdatedAt:   session.UpdatedAt,
				}
				if session.EnforcesTime() {
					summary.TimeSecs = session.TimeSecs
				}
				if session.EnforcesData() {
					summary.DataMb = session.DataMb
				}
			}
		}

		portal.SessionInfo(api, clnt, session, summary, isRunning, summary.HasAccess()).Render(r.Context(), w)
	}
}

// PortalInternetStatusCtrl re-renders the internet up/down line on internet
// up/down SSE events.
func PortalInternetStatusCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		portal.InternetStatus(api, api.Machine().IsOnline()).Render(r.Context(), w)
	}
}
