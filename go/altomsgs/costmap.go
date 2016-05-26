package altomsgs

import (
	_ "fmt"
	)

// JSON field names for CostType message fields.
const (
	FN_COST_MAP = "cost-map"
	)

// CostMap represents an ALTO Cost Map response.
// It implements the AltoMsg interface.
type CostMap struct {
	costType CostType
	depVTags []VTag
	costs map[string]map[string]Cost
}

// Verify that CostMap implements AltoMsg.
var _ AltoMsg = &CostMap{}

// NewCostMap() creates an empty cost map.
func NewCostMap() *CostMap {
	return &CostMap{
		costType: CostType{CT_ROUTINGCOST, CT_NUMERICAL},
		depVTags: make([]VTag, 0, 1),
		costs: map[string]map[string]Cost{},
	}
}

// MediaType() returns the media-type for this message.
func (this *CostMap) MediaType() string {
	return MT_COST_MAP
}

// CostType() returns the cost type for this cost map.
func (this *CostMap) CostType() CostType {
	return this.costType
}

// SetCostType() sets the cost type for this cost map.
func (this *CostMap) SetCostType(costType CostType) {
	this.costType = costType
}

// DepVTag() returns the VTag of the first resource
// upon which this cost map depends, or an empty VTag
// if there are no dependent resources.
func (this *CostMap) DepVTag() VTag {
	if this.depVTags != nil && len(this.depVTags) >= 1 {
		return this.depVTags[0]
	} else {
		return VTag{}
	}
}

// DepVTags() returns a slice with the VTags for all resources
// upon which this cost map depends.
// If none, it returns an empty array.
func (this *CostMap) DepVTags() []VTag {
	if this.depVTags == nil {
		this.depVTags = make([]VTag, 0, 1)
	}
	return this.depVTags
}

// AddDepVTag() adds a dependent VTag to this cost map.
func (this *CostMap) AddDepVTag(vtag VTag) {
	if this.depVTags == nil {
		this.depVTags = make([]VTag, 0, 1)
	}
	this.depVTags = append(this.depVTags, vtag)
}

// SetCost() sets the cost from a source to destination pid.
func (this *CostMap) SetCost(src, dst string, cost Cost) {
	if this.costs == nil {
		this.costs = map[string]map[string]Cost{}
	}
	srcmap, exists := this.costs[src]
	if !exists {
		srcmap = map[string]Cost{}
		this.costs[src] = srcmap
	}
	srcmap[dst] = cost
}

// GetCosts returns a map from source pids to destination pids to costs.
// Callers SHOULD NOT modify the retured map.
func (this *CostMap) GetCosts() map[string]map[string]Cost {
	return this.costs
}

// GetCost() returns the cost from a source to a destination pid.
// Return false if the cost map does not have that cost.
func (this *CostMap) GetCost(src, dst string) (Cost, bool) {
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
func (this *CostMap) CostIter(f func(src, dst string, cost Cost) bool) bool {
	for src, srccosts := range this.costs {
		for dst, cost := range srccosts {
			if !f(src, dst, cost) {
				return false
			}
		}
	}
	return true
}

// SrcIter() calls f(src,map[string]Cost) for all source pids in this cost map.
// The map argument gives the costs from src to each destination pid.
// If f() returns false, SrcIter() stops and returns false.
// Otherwise SrcIter() returns true after calling f() on all source pids.
func (this *CostMap) SrcIter(f func(src string, costs map[string]Cost) bool) bool {
	for src, srcmap := range this.costs {
		if !f(src, srcmap) {
			return false
		}
	}
	return true
}

// AllSrcs() returns all source PIDs used in a cost map.
func (this *CostMap) AllSrcs() []string {
	srcs := make([]string, 0, len(this.costs))
	for src, _ := range this.costs {
		srcs = append(srcs, src)
	}
	return srcs
}

// AllDsts() returns all destination PIDs used in a cost map.
func (this *CostMap) AllDsts() []string {
	dstMap := make(map[string]bool, len(this.costs))
	for _, dsts := range this.costs {
		for dst, _ := range dsts {
			dstMap[dst] = true
		}
	}
	dsts := make([]string, 0, len(dstMap))
	for k, _ := range dstMap {
		dsts = append(dsts, k)
	}
	return dsts
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
// The created map has a pointer to the data in this structure,
// especially the cost matrix, rather than a deep copy.
// Hence you must not change the cost data after calling this function.
func (this *CostMap) ToJsonMap() JsonMap {
	jm := JsonMap{}
	jm.SetCostType(this.costType)
	for _, vtag := range this.depVTags {
		jm.AddDepVTag(vtag)
	}
	jm[FN_COST_MAP] = this.costs
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
// This function uses a pointer to the map data,
// especially the cost matrix, rather than making a deep copy.
// Hence you must not change the map after calling this function.
func (this *CostMap) FromJsonMap(jm JsonMap) []error {
	errors := make([]error, 0)
	this.SetCostType(jm.GetCostType())
	for _, vtag := range jm.GetDepVTags() {
		this.AddDepVTag(vtag)
	}
	cm, ok := jm[FN_COST_MAP].(map[string]interface{})
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
									Path: FN_COST_MAP + "." + src + "." + dst,
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

