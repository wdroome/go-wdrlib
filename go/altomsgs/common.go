package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	"io"
	"fmt"
	"encoding/json"
	"bytes"
	"strings"
	"errors"
)

// Media types.
const (
	MT_PREFIX = "application/alto-"
	MT_SUFFIX = "+json"
	
	MT_DIRECTORY = MT_PREFIX + "directory" + MT_SUFFIX
	MT_NETWORK_MAP = MT_PREFIX + "networkmap" + MT_SUFFIX
	MT_NETWORK_MAP_FILTER = MT_PREFIX + "networkmapfilter" + MT_SUFFIX
	MT_COST_MAP = MT_PREFIX + "costmap" + MT_SUFFIX
	MT_COST_MAP_FILTER = MT_PREFIX + "costmapfilter" + MT_SUFFIX
	MT_ENDPOINT_COST = MT_PREFIX + "endpointcost" + MT_SUFFIX
	MT_ENDPOINT_COST_PARAMS = MT_PREFIX + "endpointcostparams" + MT_SUFFIX
	MT_ENDPOINT_PROP = MT_PREFIX + "endpointprop" + MT_SUFFIX
	MT_ENDPOINT_PROP_PARAMS = MT_PREFIX + "endpointpropparams" + MT_SUFFIX
	MT_ERROR = MT_PREFIX + "error" + MT_SUFFIX
	)

// Common JSON field names.
// Field names which are unique to a specific message type
// are defined with that message type.
const (
	FN_META = "meta"
	FN_COST_TYPE = "cost-type"
	FN_COST_METRIC = "cost-metric"
	FN_COST_MODE = "cost-mode"
	FN_VTAG = "vtag"
	FN_DEPENDENT_VTAGS = "dependent-vtags"
	FN_RESOURCE_ID = "resource-id"
	FN_TAG = "tag"
	FN_ADDRESS_TYPES = "address-types"
	FN_CONSTRAINTS = "constraints"
	FN_PIDS = "pids"
	FN_SRCS = "srcs"
	FN_DSTS = "dsts"
	FN_ENDPOINTS = "endpoints"
	FN_PROPERTIES = "properties"
	)

// CostType is an ALTO cost type.
type CostType struct {
	Metric string
	Mode string
}

// CostTypeDescription is an ALTO cost type plus a description.
type CostTypeDescription struct {
	CostType
	Description string
}

// Standard cost metrics and cost modes.
const (
	CT_ROUTINGCOST = "routingcost"
	CT_HOPCOUNT = "hopcount"
	CT_NUMERICAL = "numerical"
	CT_ORDINAL = "ordinal"
)

// String() returns a string representation of a CostType.
func (this CostType) String() string {
	return "(" + this.Metric + "/" + this.Mode + ")"
}

// String() returns a string representation of a CostTypeDescr.
func (this CostTypeDescription) String() string {
	return "(" + this.Metric + "/" + this.Mode + "/" + this.Description + ")"
}

// CostTypeListContains(list,ct) returns true iff ct equals a CostType in list.
// Return false if list is nil.
func CostTypeListContains(list []CostType, ct CostType) bool {
	if list != nil {
		for _, elem := range list {
			if ct == elem {
				return true
			}
		}
	}
	return false
}

// CostTypeListContainsAll(list,cts) returns true iff every
// CostType in cts equals a CostType in list.
// If cts is nil or 0-length, always return true.
// Otherwise, return false if list is nil.
func CostTypeListContainsAll(list []CostType, cts []CostType) bool {
	for _, ct := range cts {
		if !CostTypeListContains(list, ct) {
			return false
		}
	}
	return true
}

// CostTypeListEqual() returns true iff two slices have the same CostTypes
// in the same order. An empty slice is equal to a nil slice.
func CostTypeListEqual(a, b []CostType) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ax := range a {
		if ax != b[i] {
			return false
		}
	}
	return true
}

// CostTypeSetEqual() returns true iff two slices have the same set
// of CostTypes, possibly in different order.
func CostTypeSetEqual(a, b []CostType) bool {
	return CostTypeListContainsAll(a, b) && CostTypeListContainsAll(b, a)
}

// VTag is an ALTO version tag, with a resource id and a tag.
type VTag struct {
	ResourceId string
	Tag string
}

// Cost is the type for ALTO costs.
type Cost float32


// JsonMap is a Go map representing a JSON object (dictionary).
// The keys are the JSON field names.
type JsonMap map[string]interface{}

// AltoMsg has common functions for all ALTO message types.
type AltoMsg interface {

	// MediaType() returns the media-type for this message.
	MediaType() string
	
	// ToJsonMap() returns a map with the JSON fields
	// for the data in this structure.
	ToJsonMap() JsonMap
	
	// FromJsonMap() copies the JSON fields in a map into this structure.
	// It returns an array with the errors encountered.
	// If okay, it returns 0-length array.
	FromJsonMap(jm JsonMap) []error
}

// NewAltoMsg() reads & parses JSON from a reader and an AltoMsgwith the content.
// The function returns an array with any errors encountered;
// if there are no errors, it returns a 0-length array
// mediaType defines the message type, and must be one of the MT_* codes.
// If contentLen > 0, read at most that many bytes from r.
func NewAltoMsg(mediaType string, r io.Reader, contentLen int) (AltoMsg, []error) {
	var msg AltoMsg
	switch mediaType {
	case MT_DIRECTORY:
		msg = NewDirectory()
	case MT_NETWORK_MAP:
		msg = NewNetworkMap()
	case MT_NETWORK_MAP_FILTER:
		msg = NewNetworkMapFilter()
	case MT_COST_MAP:
		msg = NewCostMap()
	case MT_COST_MAP_FILTER:
		msg = NewCostMapFilter()
	case MT_ENDPOINT_COST:
		msg = NewEndpointCost()
	case MT_ENDPOINT_COST_PARAMS:
		msg = NewEndpointCostParams()
	case MT_ENDPOINT_PROP:
		msg = NewEndpointProp()
	case MT_ENDPOINT_PROP_PARAMS:
		msg = NewEndpointPropParams()
	case MT_ERROR:
		msg = NewErrorResp("")
	default:
		return nil, []error{errors.New("Unknown media type \"" + mediaType + "\"")}
	}
	if contentLen > 0 {
		r = &io.LimitedReader{R: r, N: int64(contentLen)}
	}
	errs := ReadJson(msg, r)
	return msg, errs
}

// ToJsonBytes() encodes the data in this structure into a JSON message.
// If successful it returns a []byte with the JSON data;
// if not, it returns a non-nil errror.
func ToJsonBytes(msg AltoMsg) ([]byte, error) {
	jm := msg.ToJsonMap()
	bytes, err := json.Marshal(&jm)
	return bytes, err
}

// WriteJson() writes the data in this structure as a JSON message.
// If successful it returns nil; if not, it returns an errror.
func WriteJson(msg AltoMsg, w io.Writer) error {
	enc := json.NewEncoder(w)
	jm := msg.ToJsonMap()
	return enc.Encode(&jm)
}

// FromJsonBytes() decodes a JSON message and copies the data
// into this structure.
// It returns an array with the errors encountered.
// If okay, it returns 0-length array.
func FromJsonBytes(msg AltoMsg, b []byte) []error {
	jm := JsonMap{}
	err := json.Unmarshal(b, &jm)
	if err != nil {
		return []error{err}
	}
	return msg.FromJsonMap(jm)
}

// ReadJson() reads and decodes a JSON input stream
// and copies the data into this structure.
// It returns an array with the errors encountered.
// If okay, it returns 0-length array.
// Caveat: The method MAY read past the end of the JSON data.
func ReadJson(msg AltoMsg, r io.Reader) []error {
	dec := json.NewDecoder(r)
	var jm JsonMap
	if err := dec.Decode(&jm); err != nil {
		// fmt.Println("Decode error:", err)
		return []error{err}
	}
	return msg.FromJsonMap(jm)
}

// PrintAltoMsg() prints the JSON for an Alto message
// in a consistent, repeatable fashion.
// It returns an error if ToJsonBytes() cannot create the JSON
// for the messaage, or if the created JSON is not parsable.
func PrintAltoMsg(msg AltoMsg, w io.Writer) {
	json, err := ToJsonBytes(msg)
	if err != nil {
		fmt.Fprintln(w, "Error creating json:", err.Error())
	} else if wdrlib.PrintJson(json, w) != nil {
		fmt.Fprintln(w, "Error re-reading created json:", err.Error())
	}
}

// CmpAltoMsgs() compares two ALTO messages by comparing their JSON strings.
// Return "" if the messages are the same.
// If not, return a description of the difference.
func CmpAltoMsgs(m1, m2 AltoMsg) (errMsg string) {
	json1, err := ToJsonBytes(m1)
	if err != nil {
		return "Error creating json for m1: " + err.Error()
	}
	buff1 := bytes.Buffer{}
	if wdrlib.PrintJson(json1, &buff1) != nil {
		return "Error creating json for m1: " + err.Error()
	}
	json2, err := ToJsonBytes(m2)
	if err != nil {
		return "Error creating json for m2: " + err.Error()
	}
	buff2 := bytes.Buffer{}
	if wdrlib.PrintJson(json2, &buff2) != nil {
		return "Error creating json for m2: " + err.Error()
	}

	nlines := 0
	for {
		nlines++
		line1, err1 := buff1.ReadString('\n')
		line1 = strings.TrimSuffix(line1, "\n")
		line2, err2 := buff2.ReadString('\n')
		line2 = strings.TrimSuffix(line2, "\n")
		if line1 != line2 {
			return fmt.Sprintf("Differ line %d: \"%s\" \"%s\"",
						nlines, line1, line2)
		} else if false {
			fmt.Printf("Same line %d: \"%s\" \"%s\"\n",
						nlines, line1, line2)
		}
		if err1 != nil && err2 != nil {
			return ""
		} else if err1 != nil {
			return "EOF on m1"
		} else if err2 != nil {
			return "EOF on m2"
		}
	}
	return ""
}
