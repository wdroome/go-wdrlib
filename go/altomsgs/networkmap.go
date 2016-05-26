package altomsgs

import (
	"net"
	"io"
	"sort"
	"fmt"
	)

// JSON field names for NetworkMap message fields.
const (
	FN_NETWORK_MAP = "network-map"
	)

// CoatMap represents an ALTO NetworkMap response.
// It implements the AltoMsg interface.
type NetworkMap struct {
	// The VTag for this map.
	vtag VTag
	
	// pids2cidrs is a map from pids => addrtypes => cidr strings.
	pids2cidrs map[string]map[string][]string
	
	// cidrsByLen is a map from each distinct cidr length to an array
	// of cidr/pid assignments for cidrs of that length.
	// cidrsByLen and pids2cidrs have the same data,
	// just organized differently. AddCIDR() keeps them in sync. 
	cidrsByLen map[int][]CIDRInfo
	
	// maskLengths has the keys in cidrsByLen, largest first.
	maskLengths []int
}

// Verify that Network implements AltoMsg.
var _ AltoMsg = &NetworkMap{}

// CIDRInfo describes a CIDR that has been assigned to a PID in a network map.
type CIDRInfo struct {
	// Pid is the PID for this CIDR.
	Pid string
	// Ipnet is the address and mask for the CIDR.
	Ipnet net.IPNet
	// MaskLen is the length of the CIDR mask.
	MaskLen int
}

// NewNetworkMap() creates an empty network map.
func NewNetworkMap() *NetworkMap {
	return &NetworkMap{
		vtag: VTag{},
		pids2cidrs: map[string]map[string][]string{},
		cidrsByLen: map[int][]CIDRInfo{},
		maskLengths: []int{},
	}
}

// makeFields() creates any maps or arrays which are nil.
func (this *NetworkMap) makeFields() {
	if this.pids2cidrs == nil {
		this.pids2cidrs = map[string]map[string][]string{}
	}
	if this.cidrsByLen == nil {
		this.cidrsByLen = map[int][]CIDRInfo{}
	}
	if this.maskLengths == nil {
		this.maskLengths = []int{}
	}
}

// MediaType() returns the media-type for this message.
func (this *NetworkMap) MediaType() string {
	return MT_NETWORK_MAP
}

// VTag() returns the VTag for this network map.
func (this *NetworkMap) VTag() VTag {
	return this.vtag
}

// SetVTag() sets the VTag for this network map.
func (this *NetworkMap) SetVTag(vtag VTag) {
	this.vtag = vtag
}

// SetAddrType() assigns a list of CIDRs to a pid.
// This returns an []error for any errors encountered.
// If there are no errors, this returns a 0-length array.
// Error include badly formatted CIDRs, CIDRs with an address
// type other than addrType, and CIDRs previously assigned
// to another PID. This function will assign the valid CIDRs
// to the PID, even if there are errors.
func (this *NetworkMap) AddCIDRs(pid, addrtype string, cidrs []string) []error {
	errors := make([]error, 0)
	for _, cidr := range cidrs {
		err := this.AddCIDR(pid, addrtype, cidr)
		if err != nil {
			errors = append(errors, err)
		}
	}
	return errors
}

// AddCIDR() assigns a CIDR to a PID.
// Return an error if cidr is invalid, or is not of addrType,
// or has been assigned to a different pid.
func (this *NetworkMap) AddCIDR(pid, addrType, cidr string) error {
	this.makeFields()

	// Validate the cidr and convert to cannonical string.
	_, ipnet, err := net.ParseCIDR(cidr)
	if err != nil {
		return CIDRError{CIDR: cidr, Err: err.Error()}
	}
	addrTypeLen, err := AddrTypeLen(addrType)
	if err != nil {
		return err
	}
	if addrTypeLen != len(ipnet.IP) {
		return CIDRError{CIDR: cidr, Err: "Not type " + addrType}
	}
	maskLen, _ := ipnet.Mask.Size()
	cidr = ipnet.String()
	
	// Create a CIDRInfo and add to cidrsByLen[].
	cidrInfo := CIDRInfo{
					Ipnet: *ipnet,
					Pid: pid,
					MaskLen: maskLen,
				}
	cidrArr, ok := this.cidrsByLen[maskLen]
	recalcLensUsed := false
	if !ok {
		cidrArr = make([]CIDRInfo, 0, 10)
		recalcLensUsed = true
	}
	for _, xcidr := range cidrArr {
		if cidrInfo.MaskLen == maskLen && cidrInfo.Ipnet.IP.Equal(xcidr.Ipnet.IP) {
			// CIDR already assigned. If same pid, return quietly.
			// If different, return an error.
			if pid != xcidr.Pid {
				return CIDRError{CIDR: cidr,
								 Err: "CIDR assigned to PID '" + xcidr.Pid + "'"}
			} else {
				return nil
			}
		}
	}
	cidrArr = append(cidrArr, cidrInfo)
	this.cidrsByLen[maskLen] = cidrArr
	if recalcLensUsed {
		this.recalcLensUsed()
	}

	// Add new CIDR/PID to the pids2cidrs map.
	pidv, ok := this.pids2cidrs[pid]
	if !ok {
		pidv = make(map[string][]string)
		this.pids2cidrs[pid] = pidv
	}
	cidrs, ok := pidv[addrType]
	if !ok {
		cidrs = make([]string, 0, 5)
	}
	for _, xcidr := range cidrs {
		if cidr == xcidr {
			return nil
		}
	}
	cidrs = append(cidrs, cidr)
	pidv[addrType] = cidrs
	return nil
}

// PidAddrs() returns the address type map for "pid".
// The keys are address types, and the values are arrays of CIDR strings.
// If pid does not exist, return false for ok.
func (this *NetworkMap) PidAddrs(pid string) (cidrMap map[string][]string, ok bool) {
	this.makeFields()
	cidrMap, ok = this.pids2cidrs[pid]
	return
}

// IP2Pid() returns the PID for an IP address,
// and the longest CIDR which contains that address.
// "ok" is true if some PID contains the address,
// or false if that address is not in any PID.
func (this *NetworkMap) IP2Pid(addr net.IP) (pid string, cidr net.IPNet, ok bool) {
	f := func(cidrInfo *CIDRInfo) bool {
		if cidrInfo.Ipnet.Contains(addr) {
			pid = cidrInfo.Pid
			cidr = cidrInfo.Ipnet
			ok = true
			return false
		}
		return true
	}
	this.CIDRIter(f)
	return
}

// CIDRIter() calls f() on all CIDRs in this network map,
// starting with the longest CIDR.
// If f() returns false, CIDRIter() stops and returns false.
// f() MUST NOT change the CIDR information.
func (this *NetworkMap) CIDRIter(f func(cidrInfo *CIDRInfo) bool) bool {
	for _, maskLength := range this.maskLengths {
		for _, cidrInfo := range this.cidrsByLen[maskLength] {
			if !f(&cidrInfo) {
				return false
			}
		}
	}
	return true
}

// PidIter() calls f(pid, addrtype, addr) for all pids and addresses.
// If f() returns false, PidIter() stops and returns false.
func (this *NetworkMap) PidIter(f func(pid, addrtype, cidr string) bool) bool {
	for pid, addrtypes := range this.pids2cidrs {
		for addrtype, cidrs := range addrtypes {
			for _, cidr := range cidrs {
				if !f(pid, addrtype, cidr) {
					return false
				}
			}
		}
	}
	return true
}

// ToJsonMap() returns a map with the JSON fields
// for the data in this structure.
// The created map has a pointer to the data in this structure,
// rather than a deep copy. Hence you must not change
// the map data after calling this function.
func (this *NetworkMap) ToJsonMap() JsonMap {
	jm := JsonMap{}
	jm.SetVTag(this.vtag)
	jm[FN_NETWORK_MAP] = this.pids2cidrs
	return jm
}

// FromJsonMap() copies the JSON fields in a map into this structure.
// It returns an array with the errors encountered.
// If okay, it returns 0-length array.
func (this *NetworkMap) FromJsonMap(jm JsonMap) []error {
	errors := []error{}
	this.makeFields()
	this.SetVTag(jm.GetVTag())
	nm, ok := jm[FN_NETWORK_MAP].(map[string]interface{})
	if ok {
		for pid, pidv := range nm {
			addrtypes, ok := pidv.(map[string]interface{})
			if ok {
				for addrtype, addrtypev := range addrtypes {
					cidrs, ok := addrtypev.([]interface{})
					if ok {
						for _, cidrv := range cidrs {
							cidr, ok := cidrv.(string)
							if ok {
								err := this.AddCIDR(pid, addrtype, cidr)
								if err != nil {
									errors = append(errors, err)
								}
							}
						}
					}
				}
			}
		}
	}
	return errors
}

// Print() writes a nicely formatted version of this network map to w.
// Note: The output is not repeatable, because the PID <=> CIDR maps
// are printed in GO's map-traversal order, which is unpredictable.
func (this *NetworkMap) Print(w io.Writer) {
	fmt.Fprintf(w, "vtag: %v\n", this.vtag)
	for _, maskLength := range this.maskLengths {
		fmt.Fprintf(w, "/%d:\n", maskLength)
		n := 0
		for _, cidrInfo := range this.cidrsByLen[maskLength] {
			if (n%5) == 0 {
				if n > 0 {
					fmt.Fprintf(w, "\n")
				}
				fmt.Fprintf(w, "   ")
			}
			fmt.Fprintf(w, "  %s: %s", cidrInfo.Ipnet.String(), cidrInfo.Pid)
			n++
		}
		if n > 0 {
			fmt.Fprintf(w, "\n")
		}
	}
	for pid, addrTypes := range this.pids2cidrs {
		fmt.Fprintf(w, "%s:\n", pid)
		for addrType, cidrs := range addrTypes {
			n := 0
			for _, cidr := range cidrs {
				if (n%6) == 0 {
					if n == 0 {
						fmt.Fprintf(w, "   %s: ", addrType)
					} else {
						fmt.Fprintf(w, "\n      ")
					}
				}
				fmt.Fprintf(w, " %s", cidr)
				n++
			}
			if n > 0 {
				fmt.Fprintf(w, "\n")
			}
		}
	}
}

// recalcLensUsed() sets maskLengths to the mask lengths, in descending order.
func (this *NetworkMap) recalcLensUsed() {
	this.maskLengths = make([]int, len(this.cidrsByLen))
	i := 0
	for k := range this.cidrsByLen {
		this.maskLengths[i] = k
		i++
	}
	sort.Sort(sort.Reverse(sort.IntSlice(this.maskLengths)))
}
		

