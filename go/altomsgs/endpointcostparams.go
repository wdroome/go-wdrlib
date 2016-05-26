package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// EndpointCostParams represents an ALTO EndpointCost request.
// It implements the AltoMsg interface.
type EndpointCostParams struct {
	// CostType is the cost type requested.
	CostType CostType
	
	// Srcs has the requested source addresses.
	// nil or 0-length means use the client's address.
	Srcs []string
	
	// Dsts has the requested destination PIDs.
	// nil or 0-length means use the client's address.
	Dsts []string
	
	// Constraints is a list of constraints.
	// nil or 0-length means no cnstraints.
	Constraints []string
}

// Verify that EndpointCostParams implements AltoMsg.
var _ AltoMsg = &EndpointCostParams{}

// NewEndpointCostParams creates a new EndpointCostParams message.
func NewEndpointCostParams() *EndpointCostParams {
	// The default init values are acceptable.
	return &EndpointCostParams{}
}

// MediaType() returns the media-type for this message.
func (this *EndpointCostParams) MediaType() string {
	return MT_ENDPOINT_COST_PARAMS
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
func (this *EndpointCostParams) ToJsonMap() JsonMap {
	jm := JsonMap{}
	jm[FN_COST_TYPE] = map[string]interface{} {
			FN_COST_METRIC: this.CostType.Metric,
			FN_COST_MODE: this.CostType.Mode,
		}
	addrs := make(map[string]interface{})
	jm[FN_ENDPOINTS] = addrs
	if this.Srcs != nil && len(this.Srcs) > 0 {
		addrs[FN_SRCS] = this.Srcs
	}
	if this.Dsts != nil && len(this.Dsts) > 0 {
		addrs[FN_DSTS] = this.Dsts
	}
	if this.Constraints != nil && len(this.Constraints) > 0 {
		jm[FN_CONSTRAINTS] = this.Constraints
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
func (this *EndpointCostParams) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	ct, ok := jm[FN_COST_TYPE].(map[string]interface{})
	if ok {
		this.CostType = CostType{
					Metric: wdrlib.GetStringMember(ct, FN_COST_METRIC),
					Mode: wdrlib.GetStringMember(ct, FN_COST_MODE),
				}
	}
	addrs, ok := jm[FN_ENDPOINTS].(map[string]interface{})
	if ok {
		this.Srcs= wdrlib.GetStringArray(addrs, FN_SRCS, nil)
		this.Dsts = wdrlib.GetStringArray(addrs, FN_DSTS, nil)
	}
	this.Constraints = wdrlib.GetStringArray(jm, FN_CONSTRAINTS, nil)
	return
}
