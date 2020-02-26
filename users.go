package main

import (
	"encoding/xml"
	"fmt"
	"strings"
)

//SearchCriteria is a search criteria for a user
type SearchCriteria struct {
	Mode          searchMode
	Field         searchField
	Value         string
	CaseSensitive bool
}

// UserGetListInSystemRequest a request to the user directory
type UserGetListInSystemRequest struct {
	XMLName  xml.Name `xml:"command"`
	Type     string   `xml:"type,attr"`
	Criteria []SearchCriteria
}

// UserGetRequest a basic request with the userId as only attribute
type UserGetRequest struct {
	XMLName xml.Name `xml:"command"`
	Type    string   `xml:"type,attr"`
	UserID  string   `xml:"userId"`
}

type searchMode string

const (
	// SearchModeStartsWith mode indicate the field should start with value
	SearchModeStartsWith searchMode = "Starts With"
	// SearchModeEquals mode indicate the field should equals to value
	SearchModeEquals searchMode = "Equals"
	// SearchModeContains mode indicate the field should contains the value
	SearchModeContains searchMode = "Contains"
)

type searchField string

const (
	// SearchFieldUserLastName use the last name as criteria
	SearchFieldUserLastName searchField = "UserLastName"
	// SearchFieldUserFirstName use the last name as criteria
	SearchFieldUserFirstName searchField = "UserFirstName"
	// SearchFieldUserID use the last name as criteria
	SearchFieldUserID searchField = "UserId"
)

//NewUserGetListInSystemRequest instantiate a new UserGetListInSystemRequest
func NewUserGetListInSystemRequest() UserGetListInSystemRequest {
	return UserGetListInSystemRequest{Type: "UserGetListInSystemRequest", Criteria: make([]SearchCriteria, 0)}
}

// Prepare returns a string ready to be to represent the request
func (r UserGetListInSystemRequest) Prepare() string {
	criteria := make([]string, len(r.Criteria))
	for i, c := range r.Criteria {
		criteria[i] = c.Prepare()
	}
	return fmt.Sprintf(`<command xmlns="" xsi:type="UserGetListInSystemRequest">
	%s
	</command>`, strings.Join(criteria, "\n"))
}

// NewSearchCriteria instantiate a new SearchCriteria
func NewSearchCriteria(mode searchMode, field searchField, value string, caseSensitive bool) (*SearchCriteria, error) {
	return &SearchCriteria{Mode: mode, Field: field, Value: value, CaseSensitive: caseSensitive}, nil
}

// Prepare returns a string ready to be to represent the request
func (c SearchCriteria) Prepare() string {
	return fmt.Sprintf(`<searchCriteria%s>
	<mode>%s</mode>
	<value>%s</value>
	<isCaseInsensitive>%v</isCaseInsensitive>
	</searchCriteria%s>`, c.Field, c.Mode, c.Value, c.CaseSensitive, c.Field)
}

// NewUserGetRequest instantiate a new UserGetRequest
func NewUserGetRequest(name, userID string) UserGetRequest {
	return UserGetRequest{Type: name, UserID: userID}
}

// Prepare returns a string ready to be to represent the request
func (r UserGetRequest) Prepare() string {
	return fmt.Sprintf(`<command xmlns="" xsi:type="%s">
    <userId>%s</userId>
</command>`, r.Type, r.UserID)
}
