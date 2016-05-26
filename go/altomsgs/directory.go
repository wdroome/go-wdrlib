package altomsgs

/*
 * One ALTO Information Resource Directory (IRD) message.
 *
 * ResourceSet describes the collection of resources
 * provided by an ALTO server.
 */

import (
	"github.com/wdroome/go/wdrlib"
	_ "fmt"
	)

// Field names for IRD messages.
const (
	FN_COST_TYPES = "cost-types"
	FN_DEFAULT_ALTO_NETWORK_MAP = "default-alto-network-map"
	FN_DESCRIPTION = "description"
	FN_RESOURCES = "resources"
	FN_URI = "uri"
	FN_MEDIA_TYPE = "media-type"
	FN_ACCEPTS = "accepts"
	FN_USES = "uses"
	FN_CAPABILITIES = "capabilities"
	FN_COST_CONSTRAINTS = "cost-constraints"
	FN_COST_TYPE_NAMES = "cost-type-names"
	FN_PROP_TYPES = "prop-types"
)

// Directory represents an ALTO Information Resource Directory (IRD) response.
// It implements the AltoMsg interface.
type Directory struct {
	// CostTypes gives the cost types defined in this message.
	// The keys are the cost type names.
	CostTypes map[string]CostTypeDescription
	
	// DefNetworkMapId is the resource id of the default network map.
	// May be "".
	DefNetworkMapId string
	
	// Resources gives information for the resources in this IRD.
	// The keys are the resource IDs.
	Resources map[string]*DirResource
}

// DirResource describes a resource in an IRD.
type DirResource struct {
	// Id is the resource's unique ID.
	Id string
	
	// URI is the resource's URI, as a string.
	URI string
	
	// MediaType is the response type for this resource.
	MediaType string
	
	// Accepts is the request type for this resource.
	// If "", this is a get-mode resource.
	Accepts string
	
	// Uses has the ID's of the resources upon which this resource depends.
	// May be nil.
	Uses []string
	
	// CostTypeNames has the names of the cost types this resource can return.
	// May be nil.
	CostTypeNames []string
	
	// CostConstraints is true iff this resource accepts cost constraint tests.
	CostConstraints bool
	
	// PropTypes has the names of the properties this resource can return.
	// May be nil.
	PropTypes []string
}

// NewDirectory() creates an empty directory.
func NewDirectory() *Directory {
	return &Directory{
		CostTypes: map[string]CostTypeDescription{},
		DefNetworkMapId: "",
		Resources: map[string]*DirResource{},
	}
}

// Verify that CostMap implements AltoMsg.
var _ AltoMsg = &Directory{}

// AddResource() adds a resource to this Directory, and returns the DirResource.
// The parameters are the fields in the new resource.
func (this *Directory) AddResource(id, uri, mediaType, accepts string,
								   uses, costTypeNames, propTypes []string,
								   constraints bool) *DirResource {
	res := DirResource{
				Id: id,
				URI: uri,
				MediaType: mediaType,
				Accepts: accepts,
				Uses: uses,
				CostTypeNames: costTypeNames,
				PropTypes: propTypes,
				CostConstraints: constraints,
			}
	this.Resources[id] = &res
	return &res
}

// MediaType() returns the media-type for this message.
func (this *Directory) MediaType() string {
	return MT_DIRECTORY
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this message.
// The created map has a pointer to the data in this structure,
// rather than a deep copy.
// Hence you must not change the data after calling this function.
func (this *Directory) ToJsonMap() JsonMap {
	jm := JsonMap{}
	
	// Meta fields.
	meta := jm.GetMeta(true)
	if this.DefNetworkMapId != "" {
		(*meta)[FN_DEFAULT_ALTO_NETWORK_MAP] = this.DefNetworkMapId
	}
	costTypes := map[string]map[string]string{}
	(*meta)[FN_COST_TYPES] = costTypes;
	for name, costType := range this.CostTypes {
		ct := map[string]string{
				FN_COST_METRIC: costType.Metric,
				FN_COST_MODE: costType.Mode,
				}
		if costType.Description != "" {
			ct[FN_DESCRIPTION] = costType.Description
		}
		costTypes[name] = ct
	}
	
	// Resource list.
	resMap := make(map[string]interface{})
	jm[FN_RESOURCES] = resMap
	for id, resource := range this.Resources {
		res := make(map[string]interface{})
		resMap[id] = res
		res[FN_URI] = resource.URI
		res[FN_MEDIA_TYPE] = resource.MediaType
		if resource.Accepts != "" {
			res[FN_ACCEPTS] = resource.Accepts
		}
		if len(resource.Uses) > 0 {
			res[FN_USES] = resource.Uses
		}
		caps := make(map[string]interface{})
		if resource.CostConstraints {
			caps[FN_COST_CONSTRAINTS] = true
		}
		if len(resource.CostTypeNames) > 0 {
			caps[FN_COST_TYPE_NAMES] = resource.CostTypeNames
		}
		if len(resource.PropTypes) > 0 {
			caps[FN_PROP_TYPES] = resource.PropTypes
		}
		if len(caps) > 0 {
			res[FN_CAPABILITIES] = caps
		}
	}
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
// This function uses a pointer to the map data,
// rather than making a deep copy.
// Hence you must not change the data after calling this function.
func (this *Directory) FromJsonMap(jm JsonMap) (errors []error) {
	errors = []error{}
	var ok bool

	if this.CostTypes == nil {
		this.CostTypes = map[string]CostTypeDescription{}
	}
	if this.Resources == nil {
		this.Resources = map[string]*DirResource{}
	}
	
	// Meta fields.
	meta := jm.GetMeta(false)
	if meta != nil {
		this.DefNetworkMapId = wdrlib.GetStringMember(*meta, FN_DEFAULT_ALTO_NETWORK_MAP)
		costTypes, ok := (*meta)[FN_COST_TYPES].(map[string]interface{})
		if ok {
			for name, ct := range costTypes {
				costType, ok := ct.(map[string]interface{})
				if ok {
					v := CostTypeDescription{}
					v.Metric = wdrlib.GetStringMember(costType, FN_COST_METRIC)
					v.Mode = wdrlib.GetStringMember(costType, FN_COST_MODE)
					v.Description = wdrlib.GetStringMember(costType, FN_DESCRIPTION)
					if this.CostTypes == nil {
						this.CostTypes = map[string]CostTypeDescription{}
					}
					this.CostTypes[name] = v
				}
			}
		}
	}
	
	// Resource list.
	xresources, ok := jm[FN_RESOURCES].(map[string]interface{})
	if ok {
		for name, xresMap := range xresources {
			resMap, ok := xresMap.(map[string]interface{})
			if ok {
				res := DirResource{
							Id: name,
							URI: wdrlib.GetStringMember(resMap, FN_URI),
							MediaType: wdrlib.GetStringMember(resMap, FN_MEDIA_TYPE),
							Accepts: wdrlib.GetStringMember(resMap, FN_ACCEPTS),
							Uses: wdrlib.GetStringArray(resMap, FN_USES, nil),
							}
				xcaps, ok := resMap[FN_CAPABILITIES].(map[string]interface{})
				if ok {
					res.CostConstraints = wdrlib.GetBoolMember(xcaps, FN_COST_CONSTRAINTS, false)
					res.CostTypeNames = wdrlib.GetStringArray(xcaps, FN_COST_TYPE_NAMES, nil)
					res.PropTypes = wdrlib.GetStringArray(xcaps, FN_PROP_TYPES, nil)
				}
				this.Resources[name] = &res
			}
		}
	}
	return
}

