package altomsgs

import (
	"testing"
	"fmt"
	_ "net"
)

func TestCheckTypedAddr(test *testing.T) {
	okAddrs := []string{
				"ipv4:1.2.3.4", "",
				"ipv4:01.2.3.4", "ipv4:1.2.3.4",
				"ipv6:::", "",
				"ipv6:a:b:c::", "",
				"ipv6:1:2:0:0:0:5::", "ipv6:1:2::5:0:0",
				"ipv6:::0", "ipv6:::",
				"ipv6:A:B:C::", "ipv6:a:b:c::",
				}
	for i := 0; i < len(okAddrs); i += 2 {
		inAddr := okAddrs[i]
		expected := okAddrs[i+1]
		if expected == "" {
			expected = inAddr
		}
		outAddr, err := CheckTypedAddr(okAddrs[i])
		if err != nil {
			test.Error("Error parsing", inAddr, err)
		} else if outAddr != expected {
			test.Error("Wrong value for", inAddr,
						"got", outAddr, "expected", expected)
		} else if false {
			fmt.Println("Paased:", inAddr, "goes to", outAddr)
		}
	}
	
	badAddrs := []string{
			"1.2.3.4",
			"ipv4:1.2.3.4.5",
			"ipv4:1.2.3",
			"ipv6:ABCDE::",
			"ipv6:ABCD:1",
			"ipv6:1.2.3.4.5.6.7.8.9",
			}
	for _, inAddr := range badAddrs {
		outAddr, err := CheckTypedAddr(inAddr)
		if err == nil {
			test.Error("Bad address accepted", inAddr, "got", outAddr)
		} else if false {
			fmt.Println("Paased:", inAddr, "gave error", err)
		}
	}
	
	var addrs []string = nil
	var expect []string = nil
	nbad := 0
	for i := 0; i < len(okAddrs); i += 2 {
		addrs = append(addrs, okAddrs[i])
		expected := okAddrs[i+1]
		if expected == "" {
			expected = okAddrs[i]
		}
		expect = append(expect, expected)
	}
	for _, addr := range badAddrs {
		addrs = append(addrs, addr)
		nbad++
	}
	outAddrs, errs := CheckTypedAddrs(addrs)
	if len(errs) != nbad {
		test.Error("CheckTypedAddrs: wrong error count", len(errs), nbad, errs)
	}
	for i, addr := range outAddrs {
		if i >= len(expect) {
			test.Error("CheckTypedAddrs: unexpected good addr", i, addr)
		} else if addr != expect[i] {
			test.Error("CheckTypedAddrs: good addr mismatch", i, addr, expect[i])
		}
	}
}

func TestParseTypedAddr(test *testing.T) {
	okAddrs := []string{
			"ipv4:1.2.3.4", "1.2.3.4",
			"ipv4:01.2.3.4", "1.2.3.4",
			"ipv6:::", "::",
			"ipv6:a:b:c::", "a:b:c::",
			"ipv6:1:2:0:0:0:5::", "1:2::5:0:0",
			"ipv6:::0", "::",
			"ipv6:A:B:C::", "a:b:c::",
			"1.2.3.4", "",
			"a:b::", "",
			}
	for i := 0; i < len(okAddrs); i += 2 {
		inAddr := okAddrs[i]
		expected := okAddrs[i+1]
		if expected == "" {
			expected = inAddr
		}
		ip, err := ParseTypedAddr(okAddrs[i])
		if err != nil {
			test.Error("Error parsing", inAddr, err)
		} else if ip.String() != expected {
			test.Error("Wrong value for", inAddr,
						"got", ip.String(), "expected", expected)
		} else if false {
			fmt.Println("Paased:", inAddr, "goes to", ip.String())
		}
	}
	
	badAddrs := []string{
			"1.2.3.4",
			"ipv4:1.2.3.4.5",
			"ipv4:1.2.3",
			"ipv6:ABCDE::",
			"ipv6:ABCD:1",
			"ipv6:1.2.3.4.5.6.7.8.9",
			}
	for _, inAddr := range badAddrs {
		ip, err := ParseTypedAddr(inAddr)
		if err == nil {
			test.Error("Bad address accepted", inAddr, "got", ip.String())
		} else if false {
			fmt.Println("Paased:", inAddr, "gave error", err)
		}
	}
}
