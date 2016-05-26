package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// NetworkMapFilter represents an ALTO network map filter request.
// It implements the AltoMsg interface.
type NetworkMapFilter struct {
	// Pids has the requested PIDs.
	// nil or 0-length means return all PIDs.
	Pids []string

	// AddrTypes has the requested address types.
	// nil or 0-length means return all address types.
	AddrTypes []string
}

// Verify that NetworkMapFilter implements AltoMsg.
var _ AltoMsg = &NetworkMapFilter{}

// NewNetworkMapFilter creates a new NetworkMapFilter message.
func NewNetworkMapFilter() *NetworkMapFilter {
	// The default init values are acceptable.
	return &NetworkMapFilter{}
}

// MediaType() returns the media-type for this message.
func (this *NetworkMapFilter) MediaType() string {
	return MT_NETWORK_MAP_FILTER
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
func (this *NetworkMapFilter) ToJsonMap() JsonMap {
	jm := JsonMap{}
	if len(this.Pids) > 0 {
		jm[FN_PIDS] = this.Pids
	}
	if len(this.AddrTypes) > 0 {
		jm[FN_ADDRESS_TYPES] = this.AddrTypes
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
func (this *NetworkMapFilter) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	pids, ok := jm[FN_PIDS].([]interface{})
	if ok {
		this.Pids = wdrlib.IfaceArrToStrs(pids)
	}
	addrTypes, ok := jm[FN_ADDRESS_TYPES].([]interface{})
	if ok {
		this.AddrTypes = wdrlib.IfaceArrToStrs(addrTypes)
	}
	return
}
