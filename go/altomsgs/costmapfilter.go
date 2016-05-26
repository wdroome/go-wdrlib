package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// CoatMapFilter represents an ALTO cost map filter request.
// It implements the AltoMsg interface.
type CostMapFilter struct {
	// CostType is the cost type requested.
	CostType CostType
	
	// Srcs has the requested source PIDs.
	// nil or 0-length means use all PIDs.
	Srcs []string
	
	// Dsts has the requested destination PIDs.
	// nil or 0-length means use all PIDs.
	Dsts []string
	
	// Constraints is a list of constraints.
	// nil or 0-length means no constraints.
	Constraints []string
}

// Verify that CostMap implements AltoMsg.
var _ AltoMsg = &CostMapFilter{}

// NewCostMapFilter creates a new CostMapFilter message.
func NewCostMapFilter() *CostMapFilter {
	// The default init values are acceptable.
	return &CostMapFilter{}
}

// MediaType() returns the media-type for this message.
func (this *CostMapFilter) MediaType() string {
	return MT_COST_MAP_FILTER
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
func (this *CostMapFilter) ToJsonMap() JsonMap {
	jm := JsonMap{}
	if this.CostType.Metric != "" || this.CostType.Mode != "" {
		jm[FN_COST_TYPE] = map[string]interface{} {
				FN_COST_METRIC: this.CostType.Metric,
				FN_COST_MODE: this.CostType.Mode,
			}
	}
	pids := make(map[string]interface{})
	jm[FN_PIDS] = pids
	if this.Srcs != nil && len(this.Srcs) > 0 {
		pids[FN_SRCS] = this.Srcs
	}
	if this.Dsts != nil && len(this.Dsts) > 0 {
		pids[FN_DSTS] = this.Dsts
	}
	if this.Constraints != nil && len(this.Constraints) > 0 {
		jm[FN_CONSTRAINTS] = this.Constraints
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
func (this *CostMapFilter) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	ct, ok := jm[FN_COST_TYPE].(map[string]interface{})
	if ok {
		this.CostType = CostType{
					Metric: wdrlib.GetStringMember(ct, FN_COST_METRIC),
					Mode: wdrlib.GetStringMember(ct, FN_COST_MODE),
				}
	}
	pids, ok := jm[FN_PIDS].(map[string]interface{})
	if ok {
		this.Srcs = wdrlib.GetStringArray(pids, FN_SRCS, nil)
		this.Dsts = wdrlib.GetStringArray(pids, FN_DSTS, nil)
	}
	this.Constraints = wdrlib.GetStringArray(jm, FN_CONSTRAINTS, nil)
	return
}
