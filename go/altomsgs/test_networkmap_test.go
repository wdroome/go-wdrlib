package altomsgs

import (
	"testing"
	"strconv"
	"net"
	_ "bytes"
	_ "os"
	_ "fmt"
	)

func TestNetworkMap(test *testing.T) {
	vtag := VTag{"myid", "314159"}
	pnetmap := testMakeNetMap(test, vtag, 64, 32)
	// pnetmap.Print(os.Stdout)
	testCheckPid(test, pnetmap, "1.2.1.5", "P24_1")
	testCheckPid(test, pnetmap, "1.2.2.4", "P24_2")
	testCheckPid(test, pnetmap, "1.4.2.4", "P16_4")
	testCheckPid(test, pnetmap, "0.0.0.0", "Default")
	testCheckPid(test, pnetmap, "1:2:1::", "P24_1")
	testCheckPid(test, pnetmap, "1:2:2::", "P24_2")
	testCheckPid(test, pnetmap, "1:4::", "P16_4")
	testCheckPid(test, pnetmap, "::", "Default")
	
	pnetmap2 := testRegenNetworkMap(test, "64/32 net", pnetmap, vtag)
	testCheckPid(test, pnetmap, "1.2.1.5", "P24_1")
	testCheckPid(test, pnetmap, "1.2.2.4", "P24_2")
	testCheckPid(test, pnetmap, "1.4.2.4", "P16_4")
	testCheckPid(test, pnetmap, "0.0.0.0", "Default")
	testCheckPid(test, pnetmap, "1:2:1::", "P24_1")
	testCheckPid(test, pnetmap, "1:2:2::", "P24_2")
	testCheckPid(test, pnetmap, "1:4::", "P16_4")
	testCheckPid(test, pnetmap, "::", "Default")
	
	vtag = VTag{"myid2", "xxx314159"}
	pnetmap = testMakeNetMap(test, vtag, 4, 2)
	testCheckPid(test, pnetmap, "1.2.1.5", "P24_1")
	testCheckPid(test, pnetmap, "1.2.2.4", "P16_2")
	testCheckPid(test, pnetmap, "1.4.2.4", "Default")
	testCheckPid(test, pnetmap, "0.0.0.0", "Default")
	testCheckPid(test, pnetmap, "1:2:1::", "P24_1")
	testCheckPid(test, pnetmap, "1:2:2::", "P16_2")
	testCheckPid(test, pnetmap, "1:4::", "Default")
	testCheckPid(test, pnetmap, "::", "Default")

	pnetmap2 = testRegenNetworkMap(test, "4/2 net", pnetmap, vtag)
	testCheckPid(test, pnetmap2, "1.2.1.5", "P24_1")
	testCheckPid(test, pnetmap2, "1.2.2.4", "P16_2")
	testCheckPid(test, pnetmap2, "1.4.2.4", "Default")
	testCheckPid(test, pnetmap2, "0.0.0.0", "Default")
	testCheckPid(test, pnetmap2, "1:2:1::", "P24_1")
	testCheckPid(test, pnetmap2, "1:2:2::", "P16_2")
	testCheckPid(test, pnetmap2, "1:4::", "Default")
	testCheckPid(test, pnetmap2, "::", "Default")
	
	pnetmap = testMakeNetMap(test, vtag, 128, 64)
	testRegenNetworkMap(test, "128/64 net", pnetmap, vtag)
}

func testRegenNetworkMap(test *testing.T,
						descr string,
						pmsg *NetworkMap,
						vtag VTag) *NetworkMap {
	altomsg2 := testRegenAltoMsg(test, descr, pmsg)
	if altomsg2 == nil {
		return nil
	} else {
		pmsg2, ok := altomsg2.(*NetworkMap)
		if !ok {
			test.Error(descr, "Regen has wrong type")
			return nil
		} else {
			vtag2 := pmsg2.VTag()
			if vtag2.ResourceId != vtag.ResourceId || vtag2.Tag != vtag.Tag {
				test.Error(descr, "Regen vtag err: Got",
							vtag2, "expected", vtag)
			}
			return pmsg2
		}
	}
}

func testMakeNetMap(test *testing.T, vtag VTag, n16, n24 int) *NetworkMap {
	pnetmap := NewNetworkMap()
	pnetmap.SetVTag(vtag)
	descr := strconv.Itoa(n16) + "/" + strconv.Itoa(n24) + " net:"
	
	var err error
	err = pnetmap.AddCIDR("Default", "ipv4", "0.0.0.0/0")
	if err != nil {
		test.Error(descr, "AddCIDR 0.0.0.0/0 error: ", err)
	}		
	err = pnetmap.AddCIDR("Default", "ipv6", "::0/0")
	if err != nil {
		test.Error(descr, "AddCIDR ::0/0 error: ", err)
	}
	var pid, addr string
	for i := 0; i < n24; i++ {
		pid = "P24_" + strconv.Itoa(i % 16)
		addr = "1.2." + strconv.Itoa(i) + ".0/24"
		err = pnetmap.AddCIDR(pid, "ipv4", addr)
		if err != nil {
			test.Error(descr, "AddCIDR", pid, addr, "error: ", err)
		}
		addr = "1:2:" + strconv.Itoa(i) + "::/48"
		err = pnetmap.AddCIDR(pid, "ipv6", addr)
		if err != nil {
			test.Error(descr, "AddCIDR", pid, addr, "error: ", err)
		}
	}
	for i := 0; i < n16; i++ {
		pid = "P16_" + strconv.Itoa(i % 16)
		addr = "1." + strconv.Itoa(i) + ".0.0/16"
		err = pnetmap.AddCIDR(pid, "ipv4", addr)
		if err != nil {
			test.Error(descr, "AddCIDR", pid, addr, "error: ", err)
		}
		addr = "1:" + strconv.Itoa(i) + "::/32"
		err = pnetmap.AddCIDR(pid, "ipv6", addr)
		if err != nil {
			test.Error(descr, "AddCIDR", pid, addr, "error: ", err)
		}
	}

	ncidrs := 0
	pnetmap.CIDRIter(
		func(cidrInfo *CIDRInfo) bool {
			ncidrs++
			return true
		})
	if ncidrs != 2*(n16+n24) + 2 {
		test.Error(descr, "incorrect CIDR count",
					ncidrs, "vs", 2*(n16+n24) + 2)
	}

	return pnetmap
}

func testCheckPid(test *testing.T, pnetmap *NetworkMap, addr string, expectedPid string) {
	ip := net.ParseIP(addr)
	if ip == nil {
		test.Error("checkPid: Bad address:", addr)
	} else {
		pid, cidr, ok := pnetmap.IP2Pid(ip)
		if !ok {
			test.Error("checkPid addr not found:", addr)
		} else {
			if expectedPid != "" && pid != expectedPid {
				test.Error("checkPid ", addr, ": Expected: ", expectedPid,
							"Got: ", pid, cidr.String())
			}
		}
	}
}
