package main

import (
	"crypto/md5"
	"crypto/sha1"
	"encoding/xml"
	"fmt"
	"io"
)

// AuthenticationRequest authenticate a user
type AuthenticationRequest struct {
	XMLName xml.Name `xml:"command"`
	Type    string   `xml:"type,attr"`
	UserID  string   `xml:"userId"`
}

// NewAuthenticationRequest prepare a new AuthenticationRequest
func NewAuthenticationRequest(userID string) AuthenticationRequest {
	return AuthenticationRequest{Type: "AuthenticationRequest", UserID: userID}
}

// Prepare returns a string version ready to be sent
func (r AuthenticationRequest) Prepare() string {
	return fmt.Sprintf(`<command xmlns="" xsi:type="AuthenticationRequest">
	<userId>%s</userId>
	</command>`, r.UserID)
}

// LoginRequest authenticate a user
type LoginRequest struct {
	XMLName         xml.Name `xml:"command"`
	Type            string   `xml:"type,attr"`
	UserID          string   `xml:"userId"`
	EncodedPassword string   `xml:"signedPassword"`
}

// NewLoginRequest prepare a new LoginRequest
func NewLoginRequest(userID, password, nonce string) LoginRequest {
	h1 := sha1.New()
	io.WriteString(h1, password)
	h2 := md5.New()
	io.WriteString(h2, nonce+":"+fmt.Sprintf("%x", h1.Sum(nil)))
	return LoginRequest{Type: "LoginRequest", UserID: userID, EncodedPassword: fmt.Sprintf("%x", h2.Sum(nil))}
}

// Prepare returns a string version ready to be sent
func (r LoginRequest) Prepare() string {
	return fmt.Sprintf(`<command xmlns="" xsi:type="LoginRequest14sp4">
		<userId>%s</userId>
		<signedPassword>%s</signedPassword>
	</command>`, r.UserID, r.EncodedPassword)
}
