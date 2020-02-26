package main

import (
	"encoding/xml"
	"fmt"
)

// SystemRequest is a request with no attribute and no paramters
type SystemRequest struct {
	XMLName xml.Name `xml:"command"`
	Type    string   `xml:"type,attr"`
}

// NewSystemRequest instanciate a new SystemRequest
func NewSystemRequest(name string) SystemRequest {
	return SystemRequest{Type: name}
}

// Prepare a request ready to be sent
func (r SystemRequest) Prepare() string {
	return fmt.Sprintf(`<command xmlns="" xsi:type="%s">
	</command>`, r.Type)
}
