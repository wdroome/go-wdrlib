package altomsgs

import (
	"testing"
	"fmt"
	"math"
	"github.com/wdroome/go/wdrlib"
	_ "bytes"
	_ "os"
	)

func TestCoatMap(test *testing.T) {
	npids := 1000
	ct := CostType{"routingcost", "numerical"}
	vt := VTag{"myid", "12345"}
	pmsg := testMakeCostMap(npids)
	if pmsg.MediaType() != MT_COST_MAP {
		test.Error("NewCostMap returns wrong media type",
					pmsg.MediaType())
	}
	pmsg.SetCostType(ct)
	pmsg.AddDepVTag(vt)
	maxCost := testMaxCost(pmsg)
	if maxCost != (Cost)(npids - 1) {
		test.Error(npids, "Pid Costmap: incorrect maxCost on orig:", maxCost)
	}
	testChkCostMap(test, pmsg, npids)
	pmsg2 := testRegenCostMap(test, fmt.Sprintf("%d pid Costmap", npids), pmsg, ct, vt)
	if pmsg2 != nil {
		maxCost2 := testMaxCost(pmsg2)
		if maxCost != maxCost {
			test.Error(npids, "Pid Costmap: incorrect maxCost on regen:", maxCost2)
		}
	}
	testChkCostMap(test, pmsg2, npids)
}

func testRegenCostMap(test *testing.T,
						descr string,
						pmsg *CostMap,
						costType CostType,
						vtag VTag) *CostMap {
	altomsg2 := testRegenAltoMsg(test, descr, pmsg)
	if altomsg2 == nil {
		return nil
	} else {
		pmsg2, ok := altomsg2.(*CostMap)
		if !ok {
			test.Error(descr, "Regen has wrong type")
			return nil
		} else {
			costType2 := pmsg2.CostType()
			vtag2 := pmsg2.DepVTag()
			if costType2.Metric != costType.Metric || costType2.Mode != costType.Mode {
				test.Error(descr, "Regen costtype err: Got",
							costType2, "expected", costType)
			}
			if vtag2.ResourceId != vtag.ResourceId || vtag2.Tag != vtag.Tag {
				test.Error(descr, "Regen dep-vtag err: Got",
							vtag2, "expected", vtag)
			}
			diff := CmpAltoMsgs(pmsg, pmsg2)
			if diff != "" {
				test.Error(descr, "Regen readback diff:", diff)
			}
			return pmsg2
		}
	}
}

func testMakeCostMap(npids int) *CostMap {
	pmap := NewCostMap()
	for isrc := 1; isrc <= npids; isrc++ {
		src := fmt.Sprintf("PID_%d", isrc)
		for idst := 1; idst <= npids; idst++ {
			dst := fmt.Sprintf("PID_%d", idst)
			pmap.SetCost(src, dst,
				(Cost)(math.Abs((float64)(isrc - idst))))
		}
	}
	return pmap
}

func testChkCostMap(test *testing.T, pmap *CostMap, npids int) {
	for isrc := 1; isrc <= npids; isrc++ {
		src := fmt.Sprintf("PID_%d", isrc)
		for idst := 1; idst <= npids; idst++ {
			dst := fmt.Sprintf("PID_%d", idst)
			got, ok := pmap.GetCost(src, dst);
			if !ok {
				got = -1
			}
			expected := (Cost)(math.Abs((float64)(isrc - idst)))
			if got != expected {
				test.Error("testChkCostMap failed: got=", got,
						"expected=", expected)
			}
		}
	}
}

func testMaxCost(pmap *CostMap) Cost {
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

func TestAllSrcDst(test *testing.T) {
	cm := NewCostMap()
	srcs := []string{"A", "B", "C"}
	dsts := []string{"D", "E", "F"}
	cost := Cost(1)
	for _, src := range srcs {
		for _, dst := range dsts {
			cm.SetCost(src, dst, cost)
		}
	}
	allSrcs := cm.AllSrcs()
	if !wdrlib.StrSetEqual(allSrcs, srcs) {
		test.Error("TestAllSrcDst AllSrcs failed: got=", allSrcs)
	}
	allDsts := cm.AllDsts()
	if !wdrlib.StrSetEqual(allDsts, dsts) {
		test.Error("TestAllSrcDst AllDsts failed: got=", allDsts)
	}
}
