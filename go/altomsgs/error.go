package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// JSON field names for Error message fields.
// All are in the meta section.
const (
	FN_ERROR_CODE = "code"
	FN_ERROR_SYNTAX_ERROR = "syntax-error"
	FN_ERROR_FIELD = "field"
	FN_ERROR_VALUE = "value"
	)

// ALTO Error Codes.
const (
	ERROR_CODE_SYNTAX = "E_SYNTAX"
	ERROR_CODE_MISSING_FIELD = "E_MISSING_FIELD"
	ERROR_CODE_INVALID_FIELD_TYPE = "E_INVALID_FIELD_TYPE"
	ERROR_CODE_INVALID_FIELD_VALUE = "E_INVALID_FIELD_VALUE"
	)

// ErrorResp represents an ALTO error response.
// It implements the AltoMsg interface.
type ErrorResp struct {
	// Code is the error code. See ERROR_CODE_*.
	Code string
	
	// SyntaxError describes the json syntax error, for ERROR_CODE_SYNTAX.
	SyntaxError string
	
	// Field is the json pathname of the offending field, or "".
	Field string
	
	// Value is the value of the offending json field, or "".
	Value string
}

// Verify that CostMap implements AltoMsg.
var _ AltoMsg = &ErrorResp{}

// NewErrorResp creates a new ErrorResp message.
func NewErrorResp(code string) *ErrorResp {
	// The other default init values are acceptable.
	return &ErrorResp{Code: code}
}

// MediaType() returns the media-type for this message.
func (this *ErrorResp) MediaType() string {
	return MT_ERROR
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
func (this *ErrorResp) ToJsonMap() JsonMap {
	jm := JsonMap{}
	meta := jm.GetMeta(true)
	if this.Code != "" {
		(*meta)[FN_ERROR_CODE] = this.Code
	}
	if this.SyntaxError != "" {
		(*meta)[FN_ERROR_SYNTAX_ERROR] = this.SyntaxError
	}
	if this.Field != "" {
		(*meta)[FN_ERROR_FIELD] = this.Field
	}
	if this.Value != "" {
		(*meta)[FN_ERROR_VALUE] = this.Value
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
func (this *ErrorResp) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	meta := jm.GetMeta(false)
	if meta == nil {
		return
	}
	this.Code = wdrlib.GetStringMember(*meta, FN_ERROR_CODE)
	this.SyntaxError = wdrlib.GetStringMember(*meta, FN_ERROR_SYNTAX_ERROR)
	this.Field = wdrlib.GetStringMember(*meta, FN_ERROR_FIELD)
	this.Value = wdrlib.GetStringMember(*meta, FN_ERROR_VALUE)
	return
}
