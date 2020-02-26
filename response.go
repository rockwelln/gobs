package main

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"

	"github.com/clbanning/mxj"
	charset "golang.org/x/net/html/charset"
)

// BSResponse is a BSFT answer
type BSResponse struct {
	Body      string
	Error     bool
	ErrorText string
	parsed    mxj.Map
}

// ErrorDetails is a BSFT error response
type ErrorDetails struct {
	Code    int
	Summary string
}

// NewBSResponse build and parse a new BSResponse object
func NewBSResponse(b []byte) (*BSResponse, error) {
	mxj.XmlCharsetReader = charset.NewReaderLabel
	m, e := mxj.NewMapXml(b)
	if e != nil {
		fmt.Printf("invalid XML: %s", hex.Dump(b))
		return nil, e
	}
	r := BSResponse{
		Error:     false,
		ErrorText: "",
		Body:      string(b),
		parsed:    m,
	}
	if r.Error = r.IsError(); r.Error {
		if e, err := r.GetErrorDetails(); err != nil {
			r.ErrorText = e.Summary
		}
	}

	return &r, nil
}

// Get return an element from the body
func (r BSResponse) Get(path string) (interface{}, error) {
	return r.parsed.ValueForPath(path)
}

// GetTable interpretes a table in the broadsoft response
func (r BSResponse) GetTable(path string) ([]map[string]string, error) {
	headings, err := r.parsed.ValuesForPath(path + ".colHeading")
	if err != nil {
		return nil, err
	}
	rows, err := r.parsed.ValuesForPath(path + ".row")
	if err != nil {
		return nil, err
	}
	result := make([]map[string]string, len(rows))
	for i := range rows {
		cols, err := r.parsed.ValuesForPath(path + fmt.Sprintf(".row[%d].col", i))
		if err != nil {
			return nil, err
		}
		result[i] = make(map[string]string, len(cols))
		for j, h := range headings {
			result[i][h.(string)] = cols[j].(string)
		}
	}
	return result, nil
}

// IsError return true if this is an error response
func (r BSResponse) IsError() bool {
	v, err := r.Get("BroadsoftDocument.command")
	if err != nil {
		return false
	}
	return v.(map[string]interface{})["-type"].(string) == "c:ErrorResponse"
}

var errorCodeRe = regexp.MustCompile(`\[Error (\d+)\]`)

// GetErrorDetails parse an error response details
func (r BSResponse) GetErrorDetails() (*ErrorDetails, error) {
	summary, err := r.Get("BroadsoftDocument.command.summary")
	if err != nil {
		return nil, err
	}
	c := errorCodeRe.FindStringSubmatch(summary.(string))
	code := 0
	if len(c) != 0 {
		code, _ = strconv.Atoi(c[1])
	}
	return &ErrorDetails{Code: code, Summary: summary.(string)}, nil
}

func (e *ErrorDetails) Error() string {
	return fmt.Sprintf("BSFT error %d: %s", e.Code, e.Summary)
}
