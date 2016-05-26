package altomsgs

import (
	_ "github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// JSON field names for EndpointProp message fields.
const (
	FN_ENDPOINT_PROPERTIES = "endpoint-properties"
	)

// EndpointProp represents an ALTO EndpointProp response.
// It implements the AltoMsg interface.
type EndpointProp struct {
	depVTags []VTag
	props map[string]map[string]string
}

// Verify that EndpointProp implements AltoMsg.
var _ AltoMsg = &EndpointProp{}

// NewEndpoints() creates an empty property map.
func NewEndpointProp() *EndpointProp {
	return &EndpointProp{
		depVTags: []VTag{},
		props: map[string]map[string]string{},
	}
}

// MediaType() returns the media-type for this message.
func (this *EndpointProp) MediaType() string {
	return MT_ENDPOINT_PROP
}

// DepVTags() returns a slice with the VTags for all resources
// upon which this property map depends.
// If none, it returns an empty array.
func (this *EndpointProp) DepVTags() []VTag {
	if this.depVTags == nil {
		this.depVTags = make([]VTag, 0, 1)
	}
	return this.depVTags
}

// AddDepVTag() adds a dependent VTag to this property map.
func (this *EndpointProp) AddDepVTag(vtag VTag) {
	if this.depVTags == nil {
		this.depVTags = make([]VTag, 0, 1)
	}
	this.depVTags = append(this.depVTags, vtag)
}

// SetProp() sets a property value for an endpoint.
func (this *EndpointProp) SetProp(addr, name, value string) {
	if this.props == nil {
		this.props = map[string]map[string]string{}
	}
	addrmap, exists := this.props[addr]
	if !exists {
		addrmap = map[string]string{}
		this.props[addr] = addrmap
	}
	addrmap[name] = value
}

// GetProps returns a map from endpoints to property names to values.
// Callers SHOULD NOT modify the retured map.
func (this *EndpointProp) GetProps() map[string]map[string]string {
	return this.props
}

// GetProp() returns the value of a property for an address.
// Return false if that address does not have that property.
// Note: This matches on the addresses as strings.
// It does not recognize equivalent addresses with different string representations.
func (this *EndpointProp) GetProp(addr, name string) (string, bool) {
	addrmap, exists := this.props[addr]
	if exists {
		value, exists := addrmap[name]
		if exists {
			return value, true
		}
	}
	return "", false
}

// PropIter() calls f(addr,name,value) on all properties in this map.
// If f() returns false, PropIter() stops and returns false.
// Otherwise PropIter() returns true after calling f() on all properties.
func (this *EndpointProp) PropIter(f func(addr, name, value string) bool) bool {
	for addr, addrprops := range this.props {
		for name, value := range addrprops {
			if !f(addr, name, value) {
				return false
			}
		}
	}
	return true
}

// AddrIter() calls f(addr,map[string]string) for all  addresses in this map.
// The map argument gives the values for all properties for that address.
// If f() returns false, AddrIter() stops and returns false.
// Otherwise AddrIter() returns true after calling f() on all source pids.
func (this *EndpointProp) AddrIter(f func(addr string, addrProps map[string]string) bool) bool {
	for addr, addrProps := range this.props {
		if !f(addr, addrProps) {
			return false
		}
	}
	return true
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
// The created map has a pointer to the data in this structure,
// especially the property matrix, rather than a deep copy.
// Hence you must not change the property data after calling this function.
func (this *EndpointProp) ToJsonMap() JsonMap {
	jm := JsonMap{}
	for _, vtag := range this.depVTags {
		jm.AddDepVTag(vtag)
	}
	jm[FN_ENDPOINT_PROPERTIES] = this.props
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
// This function uses a pointer to the map data,
// especially the property matrix, rather than making a deep copy.
// Hence you must not change the map after calling this function.
func (this *EndpointProp) FromJsonMap(jm JsonMap) []error {
	errors := make([]error, 0)
	for _, vtag := range jm.GetDepVTags() {
		this.AddDepVTag(vtag)
	}
	pm, ok := jm[FN_ENDPOINT_PROPERTIES].(map[string]interface{})
	if ok {
		for addr, addrv := range pm {
			addrprops, ok := addrv.(map[string]interface{})
			if ok {
				for name, v := range addrprops {
					switch vv := v.(type) {
					case string:
						this.SetProp(addr, name, vv)
					case nil:
						// Ignore
					default:
						errors = append(errors, JSONTypeError{
									Path: FN_ENDPOINT_PROPERTIES + "." + addr + "." + name,
									Err: "Unknown property value type",
									})
						// fmt.Printf("Unknown prop ", v, "for", addr, " ", name)
					}
				}
			}
		}
	}
	return errors
}

