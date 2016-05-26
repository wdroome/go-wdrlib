package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	"net/url"
	"errors"
	"fmt"
	"io"
	)

// ResourceSet has the resources provided by an ALTO server.
// This includes the resources in the root IRD plus
// those in all the dependent IRDs (if any).
type ResourceSet struct {
	// URI is the URI of the root Information Resource Directory.
	URI string
	
	// Resources has the resources in this set.
	// The keys are the resource IDs.
	Resources map[string]*Resource

	// DefNetworkMapId is the resource id of the default network map.
	// May be "".
	DefNetworkMapId string
}

// NewResourceSet returns a new, empty ResourceSet.
func NewResourceSet() *ResourceSet {
	return &ResourceSet{Resources: map[string]*Resource{}}
}

// Resource describes a resource provided by the server.
// This is similar to DirResource, except that a Resource
// has a fully parsed URI, and the cost metrics & modes
// rather than cost type names.
type Resource struct {
	// Id is the resource's unique ID.
	Id string
	
	// URI is the resource's URI. It is never nil.
	// The URI field always has scheme and server,
	// even if the URI string in the IRD was relative.
	URI *url.URL
	
	// MediaType is the response type for this resource.
	MediaType string
	
	// Accepts is the media type of requests for this resource.
	// If "", this is a get-mode resource.
	Accepts string
	
	// Uses has the ID's of the resources upon which this resource depends.
	// May be nil.
	Uses []string
	
	// CostTypes has the cost types this resource can return.
	// May be nil.
	CostTypes []CostType
	
	// CostConstraints is true iff this resource accepts cost constraint tests.
	CostConstraints bool
	
	// PropTypes has the names of the properties this resource can return.
	// May be nil.
	PropTypes []string
}

// NewResource() returns a Resource for an entry in an IRD.
// dirURI is the parsed URI of the IRD, and is used as the context
// if the resource has a relative URI.
// costTypeDefns is the name to CostType map in the IRD, or nil.
// Return an error if the resource's URI is invalid,
// or if it has a cost type name that is not defined in the IRD.
func NewResource(dirURI *url.URL,
				 dirRes *DirResource,
				 costTypeDefns map[string]CostTypeDescription) (*Resource, error) {
	uri, err := dirURI.Parse(dirRes.URI)
	if err != nil {
		return nil, err
	}
	var costTypes []CostType
	if dirRes.CostTypeNames != nil && len(dirRes.CostTypeNames) > 0 {
		if costTypeDefns == nil {
			return nil, errors.New("Resource \"" + dirRes.Id +
							"\": No cost type names in IRD")
		}
		costTypes = []CostType{}
		for _, name := range dirRes.CostTypeNames {
			ct, ok := costTypeDefns[name]
			if !ok {
				return nil, errors.New("Resource \"" + dirRes.Id +
								"\": No cost type \"" + name + "\" in IRD")
			}
			costTypes = append(costTypes, ct.CostType)
		}
	}
	return &Resource{
				Id: dirRes.Id,
				URI: uri,
				MediaType: dirRes.MediaType,
				Accepts: dirRes.Accepts,
				Uses: dirRes.Uses,
				CostTypes: costTypes,
				CostConstraints: dirRes.CostConstraints,
				PropTypes: dirRes.PropTypes,
			}, nil
}

// Equal() returns true iff two Resources are identical.
func (this *Resource) Equal(other *Resource) bool {
	if other == nil {
		return false
	}
	return this.Id != other.Id ||
			*this.URI != *other.URI ||
			this.MediaType != other.MediaType ||
			this.Accepts != other.Accepts ||
			!wdrlib.StrSetEqual(this.Uses, other.Uses) ||
			!CostTypeSetEqual(this.CostTypes, other.CostTypes) ||
			this.CostConstraints != other.CostConstraints ||
			!wdrlib.StrSetEqual(this.PropTypes, other.PropTypes)
}

// AddResources() adds all resources in a Directory to this ResourceSet,
// and returns an array with any errors encountered.
// If there were no errors, AddResources() returns a 0-length array.
func (this *ResourceSet) AddResources(dir *Directory, dirURI *url.URL) []error {
	errs := []error{}
	if this.DefNetworkMapId == "" && dir.DefNetworkMapId != "" {
		this.DefNetworkMapId = dir.DefNetworkMapId
	}
	for _, dirRes := range dir.Resources {
		newRes, err := NewResource(dirURI, dirRes, dir.CostTypes)
		if err != nil {
			errs = append(errs, err)
		}
		if newRes != nil {
			curRes, exists := this.Resources[dirRes.Id]
			if !exists {
				this.Resources[dirRes.Id] = newRes
			} else if !newRes.Equal(curRes) {
				// Quietly ignore duplicate resource declarations.
				// Only report an error if the new resource is different.
				errs = append(errs, errors.New("Duplicate resource id \"" +
												dirRes.Id + "\""))
			}
		}
	}
	return errs
}

// FindDefNetworkMap() returns the default NetworkMap resource.
// If there is no explicit default, and there is only one NetworkMap,
// return it. If there is more than one NetworkMap, return nil.
func (this *ResourceSet) FindDefNetworkMap() *Resource {
	netmap, ok := this.Resources[this.DefNetworkMapId]
	if ok {
		return netmap
	}
	netmap = nil
	for _, res := range this.Resources {
		if res.MediaType == MT_NETWORK_MAP && res.Accepts == "" {
			if netmap != nil {
				// Has more than one NetworkMap, so there is no clear default.
				return nil
			}
			netmap = res
		}
	}
	return netmap
}

// FindNetworkMaps() returns all NetworkMap resources.
// If there are none, it returns a 0-length array.
func (this *ResourceSet) FindNetworkMaps() []*Resource {
	netmaps := []*Resource{}
	for _, res := range this.Resources {
		if res.MediaType == MT_NETWORK_MAP && res.Accepts == "" {
			netmaps = append(netmaps, res)
		}
	}
	return netmaps
}

// FindFilteredNetworkMap() returns the first FilteredNetworkMap resource
// which the NetworkMap resource netmap.
// Return nil if there is no such resource.
func (this *ResourceSet) FindFilteredNetworkMap(netmap string) *Resource {
	for _, res := range this.Resources {
		if res.MediaType == MT_NETWORK_MAP &&
					res.Accepts == MT_NETWORK_MAP_FILTER &&
					wdrlib.StrListContains(res.Uses, netmap) {
			return res
		}
	}
	return nil
}

// FindCostMap() returns the first CostMap resource
// which returns costType for the NetworkMap resource netmap.
// Return nil if there is no such resource.
func (this *ResourceSet) FindCostMap(
									netmap string,
									costType CostType) *Resource {
	for _, res := range this.Resources {
		if res.MediaType == MT_COST_MAP &&
					res.Accepts == "" &&
					wdrlib.StrListContains(res.Uses, netmap) &&
					CostTypeListContains(res.CostTypes, costType) {
			return res
		}
	}
	return nil
}

// FindFilteredCostMap() returns the first FilteredCostMap resource
// which returns costType for the NetworkMap resource netmap.
// If needConstraints is true, return the resource which accepts cost constraints.
// Return nil if there is no such resource.
func (this *ResourceSet) FindFilteredCostMap(
									netmap string,
									costType CostType,
									needConstraints bool) *Resource {
	for _, res := range this.Resources {
		if res.MediaType == MT_COST_MAP &&
					res.Accepts == MT_COST_MAP_FILTER &&
					wdrlib.StrListContains(res.Uses, netmap) &&
					CostTypeListContains(res.CostTypes, costType) &&
					(!needConstraints || res.CostConstraints) {
			return res
		}
	}
	return nil
}

// FindEndpointCost() returns the first EndpointCost resource
// which returns costType. If needConstraints is true,
// return the resource which accepts cost constraints.
// Return nil if there is no such resource.
func (this *ResourceSet) FindEndpointCost(
									costType CostType,
									needConstraints bool) *Resource {
	for _, res := range this.Resources {
		if res.MediaType == MT_ENDPOINT_COST &&
					res.Accepts == MT_ENDPOINT_COST_PARAMS &&
					CostTypeListContains(res.CostTypes, costType) &&
					(!needConstraints || res.CostConstraints) {
			return res
		}
	}
	return nil
}

// FindEndpointProp() returns the first EndpointProp resource
// which returns all the property types in propTypes.
// Return nil if there is no such resource.
func (this *ResourceSet) FindEndpointProp(propTypes []string) *Resource {
	for _, res := range this.Resources {
		if res.MediaType == MT_ENDPOINT_PROP &&
					res.Accepts == MT_ENDPOINT_PROP_PARAMS &&
					wdrlib.StrListContainsAll(res.PropTypes, propTypes) {
			return res
		}
	}
	return nil
}

// Print() prints a ResourceSet.
func (this *ResourceSet) Print(w io.Writer) {
	fmt.Fprintf(w, "ResourceSet: IRD: %s  DefNetMap: %s  Resources: %d\n",
					this.URI, this.DefNetworkMapId, len(this.Resources))
	for _, res := range this.Resources {
		res.Print(w, "  ")
	}
}

// Print() prints a Resource.
func (this *Resource) Print(w io.Writer, prefix string) {
	fmt.Fprintf(w, "%s%s:\n", prefix, this.Id)
	prefix += "   "
	fmt.Fprintf(w, "%sURI: %s\n", prefix, this.URI.String());
	fmt.Fprintf(w, "%sMediaType: %s\n", prefix, this.MediaType)
	if this.Accepts != "" {
		fmt.Fprintf(w, "%sAccepts: %s\n", prefix, this.Accepts)
	}
	if len(this.Uses) > 0 {
		fmt.Fprintf(w, "%sUses:", prefix);
		for _, v := range this.Uses {
			fmt.Fprintf(w, " %s", v)
		}
		fmt.Fprintf(w, "\n")
	}
	if len(this.CostTypes) > 0 {
		fmt.Fprintf(w, "%sCostTypes:", prefix);
		for _, v := range this.CostTypes {
			fmt.Fprintf(w, " %s", v)
		}
		fmt.Fprintf(w, "\n")
	}
	if this.CostConstraints {
		fmt.Fprintf(w, "%sCostContraints: true\n", prefix)
	}
	if len(this.PropTypes) > 0 {
		fmt.Fprintf(w, "%sPropTypes:", prefix);
		for _, v := range this.PropTypes {
			fmt.Fprintf(w, " %s", v)
		}
		fmt.Fprintf(w, "\n")
	}
}
