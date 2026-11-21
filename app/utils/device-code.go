package utils

import (
	"context"

	sdkapi "sdk/api"
)

// PortalDeviceCode returns the device code the portal shows clnt, or "" when
// clnt has no wifi credits or the code cannot be read. DeviceCode creates the
// code on first read, so a phone without wifi credits is not given one just
// by opening the portal.
func PortalDeviceCode(ctx context.Context, api sdkapi.IPluginApi, clnt sdkapi.IClientDevice, summary sdkapi.SessionSummary) string {
	if !summary.HasAccess() {
		return ""
	}
	code, err := clnt.DeviceCode(ctx)
	if err != nil {
		api.Logger().Error("coffee-theme: unable to read the device code: " + err.Error())
		return ""
	}
	return code
}
