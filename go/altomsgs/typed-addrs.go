package altomsgs

import (
	_ "github.com/wdroome/go/wdrlib"
	"net"
	"strings"
	_ "fmt"
	_ "errors"
)

// AddrType prefixes.
const (
	IPV4_ADDR_TYPE = "ipv4"
	IPV4_ADDR_PREFIX = IPV4_ADDR_TYPE + ":"
	IPV6_ADDR_TYPE = "ipv6"
	IPV6_ADDR_PREFIX = IPV6_ADDR_TYPE + ":"
	)

// AddrTypeLen() returns the length of an address type,
// or an error if the address type is invalid.
func AddrTypeLen(addrtype string) (int, error) {
	if addrtype == IPV4_ADDR_TYPE {
		return 4, nil
	} else if addrtype == IPV6_ADDR_TYPE {
		return 16, nil
	} else {
		return 0, net.InvalidAddrError(
					addrtype + " is not valid address type")
	}
}

// CheckTypedAddr() validates an endpoint address with a type prefix.
// If ok, the method returns the cannonical string version of inAddr,
// and nil for err. If inAddr is not valid, the method returns an error,
// and returns "" outAddr.
func CheckTypedAddr(inAddr string) (outAddr string, err error) {
	outAddr = ""
	err = nil
	if s4 := strings.TrimPrefix(inAddr, IPV4_ADDR_PREFIX); s4 != inAddr {
		ip := net.ParseIP(s4)
		if ip == nil {
			err = net.InvalidAddrError(inAddr + " is not a valid address")
		} else {
			outAddr = IPV4_ADDR_PREFIX + ip.String()
		}
	} else if s6 := strings.TrimPrefix(inAddr, IPV6_ADDR_PREFIX); s6 != inAddr {
		ip := net.ParseIP(s6)
		if ip == nil {
			err = net.InvalidAddrError(inAddr + " is not a valid address")
		} else {
			outAddr = IPV6_ADDR_PREFIX + ip.String()
		}
	} else {
		err = net.InvalidAddrError(inAddr + " is not a typed address")
	}
	return
}

// CheckTypedAddrs() validates & normalizes an array of typed addresses.
// It returns an array with the valid addresses,
// and an array with the errors found, if any.
// The returned "errs" slice may be nil or a 0-length.
// The returned "outAddrs" slice is never nil.
func CheckTypedAddrs(inAddrs []string) (outAddrs []string, errs []error) {
	if inAddrs == nil {
		inAddrs = []string{}
	}
	outAddrs = make([]string, 0, len(inAddrs))
	errs = nil
	for _, inAddr := range inAddrs {
		outAddr, err := CheckTypedAddr(inAddr)
		if err != nil {
			errs = append(errs, err)
		} else {
			outAddrs = append(outAddrs, outAddr)
		}
	}
	return
}

// ParseTypedAddr() returns the net.IP for a possibly typed address.
// The function returns an error of addr is not valid,
// or if the address part does not match the type prefix.
// If addr does not have a type prefix, the function imputes the type.
func ParseTypedAddr(addr string) (net.IP, error) {
	var untypedAddr string
	isIpv4 := false
	isIpv6 := false
	if strings.HasPrefix(addr, IPV4_ADDR_PREFIX) {
		untypedAddr = addr[len(IPV4_ADDR_PREFIX):]
		isIpv4 = true
	} else if strings.HasPrefix(addr, IPV6_ADDR_PREFIX) {
		untypedAddr = addr[len(IPV6_ADDR_PREFIX):]
		isIpv6 = true
	} else {
		untypedAddr = addr
	}
	ip := net.ParseIP(untypedAddr)
	if ip == nil {
		return nil, &net.AddrError{Err: "Invalid IP address", Addr: addr}
	}
	if isIpv4 {
		ip = ip.To4()
		if ip == nil {
			return nil, &net.AddrError{Err: "Invalid ipv4 address", Addr: addr}
		}
	} else if isIpv6 {
		ip = ip.To16()
		if ip == nil {
			return nil, &net.AddrError{Err: "Invalid ipv6 address", Addr: addr}
		}
	}
	return ip, nil
}
