package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// EndpointPropParams represents an ALTO EndpointProp request.
// It implements the AltoMsg interface.
type EndpointPropParams struct {
	// Properties has the request property names.
	// May be nil or 0-length.
	Properties []string
	
	// Endpoints has the requested addresses.
	// May be nil or 0-length.
	Endpoints []string
}

// Verify that EndpointPropParams implements AltoMsg.
var _ AltoMsg = &EndpointPropParams{}

// NewEndpointPropParams creates a new EndpointPropParams message.
func NewEndpointPropParams() *EndpointPropParams {
	// The default init values are acceptable.
	return &EndpointPropParams{}
}

// MediaType() returns the media-type for this message.
func (this *EndpointPropParams) MediaType() string {
	return MT_ENDPOINT_PROP_PARAMS
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
func (this *EndpointPropParams) ToJsonMap() JsonMap {
	jm := JsonMap{}
	if this.Properties != nil && len(this.Properties) > 0 {
		jm[FN_PROPERTIES] = this.Properties
	}
	if this.Endpoints != nil && len(this.Endpoints) > 0 {
		jm[FN_ENDPOINTS] = this.Endpoints
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
func (this *EndpointPropParams) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	this.Properties = wdrlib.GetStringArray(jm, FN_PROPERTIES, nil)
	this.Endpoints = wdrlib.GetStringArray(jm, FN_ENDPOINTS, nil)
	return
}
