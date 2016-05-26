package altomsgs

import (
	"testing"
	"fmt"
	"math"
	_ "bytes"
	"os"
	)

func TestEndpointCost(test *testing.T) {
	naddrs := 256
	ct := CostType{"routingcost", "numerical"}
	pmsg := testMakeEndpointCost(naddrs)
	if pmsg.MediaType() != MT_ENDPOINT_COST {
		test.Error("NewEndpointCoat returns wrong media type",
					pmsg.MediaType())
	}
	pmsg.SetCostType(ct)
	maxCost := testMaxEndpointCost(pmsg)
	if maxCost != (Cost)(naddrs - 1) {
		test.Error(naddrs, "addr Costmap: incorrect maxCost on orig:", maxCost)
	}
	testChkEndpointCost(test, pmsg, naddrs)
	pmsg2 := testRegenEndpointCost(test, fmt.Sprintf("%d addr Costmap", naddrs), pmsg, ct)
	if pmsg2 != nil {
		maxCost2 := testMaxEndpointCost(pmsg2)
		if maxCost != maxCost {
			test.Error(naddrs, "addr Costmap: incorrect maxCost on regen:", maxCost2)
		}
	}
	testChkEndpointCost(test, pmsg2, naddrs)

	pmsg2.Normalize()
	testChkEndpointCost(test, pmsg2, naddrs)
	
	if false {
		PrintAltoMsg(pmsg, os.Stdout)
		PrintAltoMsg(pmsg2, os.Stdout)
	}
}

func testRegenEndpointCost(test *testing.T,
						descr string,
						pmsg *EndpointCost,
						costType CostType) *EndpointCost {
	altomsg2 := testRegenAltoMsg(test, descr, pmsg)
	if altomsg2 == nil {
		return nil
	} else {
		pmsg2, ok := altomsg2.(*EndpointCost)
		if !ok {
			test.Error(descr, "Regen has wrong type")
			return nil
		} else {
			costType2 := pmsg2.CostType()
			if costType2.Metric != costType.Metric || costType2.Mode != costType.Mode {
				test.Error(descr, "Regen costtype err: Got",
							costType2, "expected", costType)
			}
			diff := CmpAltoMsgs(pmsg, pmsg2)
			if diff != "" {
				test.Error(descr, "Regen readback diff:", diff)
			}
			return pmsg2
		}
	}
}

func testMakeEndpointCost(naddrs int) *EndpointCost {
	pmap := NewEndpointCost()
	for isrc := 0; isrc < naddrs; isrc++ {
		src := testMakeEndAddr(isrc, false)
		for idst := 0; idst < naddrs; idst++ {
			dst := testMakeEndAddr(idst, false)
			pmap.SetCost(src, dst,
				(Cost)(math.Abs((float64)(isrc - idst))))
		}
	}
	return pmap
}

func testMakeEndAddr(i int, normal bool) string {
	if i % 2 == 0 {
		return fmt.Sprintf("ipv4:1.2.3.%d", i)
	} else if normal {
		return fmt.Sprintf("ipv6:1:2:a:%x::", i)
	} else {
		return fmt.Sprintf("ipv6:1:2:A:%X::", i)
	}
}

func testChkEndpointCost(test *testing.T, pmap *EndpointCost, naddrs int) {
	for isrc := 0; isrc < naddrs; isrc++ {
		src := testMakeEndAddr(isrc, pmap.IsNormalized())
		for idst := 0; idst < naddrs; idst++ {
			dst := testMakeEndAddr(idst, pmap.IsNormalized())
			got, ok := pmap.GetCost(src, dst);
			if !ok {
				got = -1
			}
			expected := (Cost)(math.Abs((float64)(isrc - idst)))
			if got != expected {
				test.Error("testChkCostMap failed:", src, dst, "got=", got,
						"expected=", expected)
			}
		}
	}
}

func testMaxEndpointCost(pmap *EndpointCost) Cost {
	mc := Cost(0)
	f := func(src, dst string, cost Cost) bool {
		if cost > mc {
			mc = cost
		}
		return true
	}
	pmap.CostIter(f)
	return mc
}
