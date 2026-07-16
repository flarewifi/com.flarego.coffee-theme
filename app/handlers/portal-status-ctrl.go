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
			_, isConnected = api.SessionsMgr().RunningSession(clnt)
		}

		portal.StatusBar(api, isConnected).Render(r.Context(), w)
	}
}

// PortalSessionInfoCtrl re-renders the time/data-left panel.
func PortalSessionInfoCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var summary *sdkapi.ClientSessionSummary
		var isRunning bool
		var sessionType sdkapi.SessionType

		clnt, err := api.Http().GetClientDevice(r)
		if err != nil {
			api.Logger().Error(fmt.Sprintf("coffee-theme session-info: get client device error: %v", err))
		} else {
			session, ok := api.SessionsMgr().RunningSession(clnt)
			isRunning = ok
			if ok {
				sessionType = session.Type()
				s, err := api.SessionsMgr().SessionSummary(r.Context(), clnt)
				if err != nil {
					api.Logger().Error(fmt.Sprintf("coffee-theme session-info: session summary error: %v", err))
				} else {
					summary = s
				}
			}
		}

		portal.SessionInfo(api, summary, isRunning, sessionType).Render(r.Context(), w)
	}
}

// PortalInternetStatusCtrl re-renders the internet up/down line on internet
// up/down SSE events.
func PortalInternetStatusCtrl(api sdkapi.IPluginApi) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		portal.InternetStatus(api, api.Machine().IsOnline()).Render(r.Context(), w)
	}
}
