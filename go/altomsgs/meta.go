package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
)

// GetMeta returns a map for the meta field in a JSON message.
// If create is true, it creates a map for meta if it does not exist.
// If create is false and meta is missing, it returns nil.
func (this *JsonMap) GetMeta(create bool) *map[string]interface{} {
	meta, ok := (*this)[FN_META].(map[string]interface{})
	if ok {
		return &meta
	} else if !create {
		return nil
	} else {
		meta = make(map[string]interface{})
		(*this)[FN_META] = meta
		return &meta
	}
}

// SetCostType sets the cost-type field in the meta section
// of a JSON message.
func (this *JsonMap) SetCostType(costType CostType) {
	(*this.GetMeta(true))[FN_COST_TYPE] = map[string]interface{} {
			FN_COST_METRIC: costType.Metric,
			FN_COST_MODE: costType.Mode,
		}
}

// GetCostType returns the CostType from the meta section
// of a JSON message. Missing values are "".
func (this *JsonMap) GetCostType() CostType {
	meta := this.GetMeta(false)
	if meta == nil {
		return CostType{}
	}
	ct, ok := (*meta)[FN_COST_TYPE].(map[string]interface{})
	if !ok {
		return CostType{}
	}
	return CostType{
			Metric: wdrlib.GetStringMember(ct, FN_COST_METRIC),
			Mode: wdrlib.GetStringMember(ct, FN_COST_MODE),
		}
}	

// SetVTag sets the VTag field in the meta section
// of a JSON message. This is the VTag for the resource in this message.
// Use AddDepVTag() to set dependent vtags.
func (this *JsonMap) SetVTag(vtag VTag) {
	(*this.GetMeta(true))[FN_VTAG] = map[string]interface{} {
			FN_RESOURCE_ID: vtag.ResourceId,
			FN_TAG: vtag.Tag,
		}
}

// GetVTag returns the VTag for this resource from the meta section
// of a JSON message. Missing values are "".
func (this *JsonMap) GetVTag() VTag {
	meta := this.GetMeta(false)
	if meta == nil {
		return VTag{}
	}
	vt, ok := (*meta)[FN_VTAG].(map[string]interface{})
	if !ok {
		return VTag{}
	}
	return VTag{
			ResourceId: wdrlib.GetStringMember(vt, FN_RESOURCE_ID),
			Tag: wdrlib.GetStringMember(vt, FN_TAG),
		}
}	

// AddDepVTag adds a dependent vtag to the meta section
// of a JSON message.
func (this *JsonMap) AddDepVTag(vtag VTag) {
	meta := this.GetMeta(true)
	depVTags, ok := (*meta)[FN_DEPENDENT_VTAGS].([]map[string]interface{})
	if !ok {
		depVTags = make([]map[string]interface{}, 0, 2)
	}
	depVTags = append(depVTags, map[string]interface{} {
			FN_RESOURCE_ID: vtag.ResourceId,
			FN_TAG: vtag.Tag,
		})
	(*meta)[FN_DEPENDENT_VTAGS] = depVTags
}

// GetDepVTags returns a slice with the dependent vtags
// in the meta section of a JSON message.
func (this *JsonMap) GetDepVTags() []VTag {
	meta := this.GetMeta(false)
	if meta == nil {
		return make([]VTag, 0)
	}
	arr, ok := (*meta)[FN_DEPENDENT_VTAGS].([]interface{})
	if !ok {
		return make([]VTag, 0)
	}
	vtags := make([]VTag, 0, len(arr))
	for _, v := range arr {
		vx, ok := v.(map[string]interface{})
		if ok {
			vtags = append(vtags, VTag{
						ResourceId: wdrlib.GetStringMember(vx, FN_RESOURCE_ID),
						Tag: wdrlib.GetStringMember(vx, FN_TAG),
					})
		}
	}
	return vtags
}
