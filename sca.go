package main

import (
	"encoding/xml"
	"fmt"
)

// UserSharedCallAppearanceGetEndpointRequest is a request for user SCA endpoints
type UserSharedCallAppearanceGetEndpointRequest struct {
	XMLName     xml.Name `xml:"command"`
	Type        string   `xml:"type,attr"`
	UserID      string   `xml:"userId"`
	DeviceName  string   `xml:"deviceName"`
	DeviceLevel string   `xml:"deviceLevel"`
	LinePort    string   `xml:"linePort"`
}

// NewUserSharedCallAppearanceGetEndpointRequest creates a new UserSharedCallAppearanceGetEndpointRequest
func NewUserSharedCallAppearanceGetEndpointRequest(userID, deviceName, deviceLevel, linePort string) UserSharedCallAppearanceGetEndpointRequest {
	return UserSharedCallAppearanceGetEndpointRequest{Type: "UserSharedCallAppearanceGetEndpointRequest", DeviceName: deviceName, DeviceLevel: deviceLevel, LinePort: linePort}
}

// Prepare turn the request into a string ready to be sent
func (r UserSharedCallAppearanceGetEndpointRequest) Prepare() string {
	return fmt.Sprintf(`<command xmlns="" xsi:type="UserSharedCallAppearanceGetEndpointRequest">
    <userId>%s</userId>
	<accessDeviceEndpoint>
		<accessDevice>%s</accessDevice>
		<linePort>%s</linePort>
	</accessDeviceEndpoint>
</command>`, r.UserID, r.DeviceName, r.LinePort)
}
