package altomsgs

import (
	_ "github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// JSON field names for EndpointCost message fields.
const (
	FN_ENDPOINT_COST_MAP = "endpoint-cost-map"
	)

// EndpointCost represents an ALTO EndpointCost response.
// It implements the AltoMsg interface.
type EndpointCost struct {
	costType CostType
	costs map[string]map[string]Cost
	normalized bool
}

// Verify that EndpointCost implements AltoMsg.
var _ AltoMsg = &EndpointCost{}

// NewEndpointCost() creates an empty cost map.
func NewEndpointCost() *EndpointCost {
	return &EndpointCost{
		costType: CostType{CT_ROUTINGCOST, CT_NUMERICAL},
		costs: map[string]map[string]Cost{},
		normalized: false,
	}
}

// MediaType() returns the media-type for this message.
func (this *EndpointCost) MediaType() string {
	return MT_ENDPOINT_COST
}

// CostType() returns the cost type for this cost map.
func (this *EndpointCost) CostType() CostType {
	return this.costType
}

// SetCostType() sets the cost type for this cost map.
func (this *EndpointCost) SetCostType(costType CostType) {
	this.costType = costType
}

// SetCost() sets the cost from a source to destination address.
func (this *EndpointCost) SetCost(src, dst string, cost Cost) {
	if this.costs == nil {
		this.costs = map[string]map[string]Cost{}
	}
	srcmap, exists := this.costs[src]
	if !exists {
		srcmap = map[string]Cost{}
		this.costs[src] = srcmap
	}
	srcmap[dst] = cost
	this.normalized = false
}

// Normalize() checks and normalizes all endpoint addresses.
// The function removes any invalid addresses from the cost map,
// and returns a list of the errors found.
// If all addresses are valid, the method returns nil or a 0-length slice.
// Normalize() returns immediately if the cost map has not changed
// since the previous call.
func (this *EndpointCost) Normalize() []error {
	errs := []error{}
	if !this.normalized {
		ncosts := make(map[string]map[string]Cost, len(this.costs))
		for src, srccosts := range this.costs {
			nsrc, err := CheckTypedAddr(src)
			if err != nil {
				errs = append(errs, err)
			} else {
				nsrccosts := make(map[string]Cost, len(srccosts))
				ncosts[nsrc] = nsrccosts
				for dst, cost := range srccosts {
					ndst, err := CheckTypedAddr(dst)
					if err != nil {
						errs = append(errs, err)
					} else {
						nsrccosts[ndst] = cost
					}
				}
			}
		}
		this.normalized = true
		this.costs = ncosts
	}
	return errs
}

// IsNormalized() returns true iff all endpoints have been checked and normalized.
// See Normalize().
func (this *EndpointCost) IsNormalized() bool {
	return this.normalized
}

// GetCosts returns a map from source address to destination address to costs.
// Callers SHOULD NOT modify the retured map.
func (this *EndpointCost) GetCosts() map[string]map[string]Cost {
	return this.costs
}

// GetCost() returns the cost from a source to a destination address.
// Return false if the cost map does not have that cost.
// Note: This matches on the addresses as strings.
// It does not recognize equivalent addresses with different string representations.
func (this *EndpointCost) GetCost(src, dst string) (Cost, bool) {
	srcmap, exists := this.costs[src]
	if exists {
		cost, exists := srcmap[dst]
		if exists {
			return cost, true
		}
	}
	return 0, false
}

// CostIter() calls f(src,dst,cost) on all cost points in this cost map.
// If f() returns false, CostIter() stops and returns false.
// Otherwise CostIter() returns true after calling f() on all cost points.
func (this *EndpointCost) CostIter(f func(src, dst string, cost Cost) bool) bool {
	for src, srccosts := range this.costs {
		for dst, cost := range srccosts {
			if !f(src, dst, cost) {
				return false
			}
		}
	}
	return true
}

// SrcIter() calls f(src,map[string]Cost) for all source addresses in this cost map.
// The map argument gives the costs from src to each destination address.
// If f() returns false, SrcIter() stops and returns false.
// Otherwise SrcIter() returns true after calling f() on all source pids.
func (this *EndpointCost) SrcIter(f func(src string, costs map[string]Cost) bool) bool {
	for src, srcmap := range this.costs {
		if !f(src, srcmap) {
			return false
		}
	}
	return true
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
// The created map has a pointer to the data in this structure,
// especially the cost matrix, rather than a deep copy.
// Hence you must not change the cost data after calling this function.
func (this *EndpointCost) ToJsonMap() JsonMap {
	jm := JsonMap{}
	jm.SetCostType(this.costType)
	jm[FN_ENDPOINT_COST_MAP] = this.costs
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
// This function uses a pointer to the map data,
// especially the cost matrix, rather than making a deep copy.
// Hence you must not change the map after calling this function.
func (this *EndpointCost) FromJsonMap(jm JsonMap) []error {
	errors := make([]error, 0)
	this.SetCostType(jm.GetCostType())
	cm, ok := jm[FN_ENDPOINT_COST_MAP].(map[string]interface{})
	if ok {
		for src, srcv := range cm {
			srccosts, ok := srcv.(map[string]interface{})
			if ok {
				for dst, v := range srccosts {
					switch vv := v.(type) {
					case float64:
						this.SetCost(src, dst, Cost(vv))
					case float32:
						this.SetCost(src, dst, Cost(vv))
					case int:
						this.SetCost(src, dst, Cost(vv))
					case nil:
						// Ignore
					default:
						errors = append(errors, JSONTypeError{
									Path: FN_ENDPOINT_COST_MAP + "." + src + "." + dst,
									Err: "Unknown cost type",
									})
						// fmt.Printf("Unknown cost point", v, "for", src, ":", dst)
					}
				}
			}
		}
	}
	return errors
}

