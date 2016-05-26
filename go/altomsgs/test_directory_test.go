package altomsgs

import (
	"testing"
	_ "bytes"
	_ "fmt"
	_ "os"
	)

func TestDirectory(test *testing.T) {
	dir := NewDirectory()
	if dir.MediaType() != MT_DIRECTORY {
		test.Error("NewDirectory returns wrong media type",
					dir.MediaType())
	}
	testAddCostType(dir.CostTypes, "num-rc",
				"routingcost", "numerical", "plain old routingcost")
	testAddCostType(dir.CostTypes, "num-hops",
				"hopcount", "numerical", "plain old hopcount")
	testAddCostType(dir.CostTypes, "ord-rc",
				"routingcost", "ordinal", "")
	netmapId := "mynetmap"
	dir.DefNetworkMapId = netmapId
	dir.AddResource(netmapId, "/mynetmap/netmap",
				MT_NETWORK_MAP, "",
				nil,
				nil, nil, false)
	dir.AddResource("num-rt-costmap", "/mynetmap/costs/num-rt",
				MT_COST_MAP, "",
				[]string{netmapId},
				[]string{"num-rc"}, nil, false)
	dir.AddResource("num-hops-costmap", "/mynetmap/costs/num-hops",
				MT_COST_MAP, "",
				[]string{netmapId},
				[]string{"num-hops"}, nil, false)
	dir.AddResource("ord-rc-costmap", "/mynetmap/costs/ord-rc",
				MT_COST_MAP, "",
				[]string{netmapId},
				[]string{"ord-rc"}, nil, false)
	dir.AddResource("filter-costmap", "/mynetmap/costs/filter",
				MT_COST_MAP, MT_COST_MAP_FILTER,
				[]string{netmapId},
				[]string{"num-rc", "num-hops", "ord-rc"}, nil, true)
	// fmt.Println(dir)
	// PrintAltoMsg(dir, os.Stdout)

	nres := len(dir.Resources)
	if nres != 5 {
		test.Error("Incorrect resource count in orig:", nres)
	}
	nct := len(dir.CostTypes)
	if nct != 3 {
		test.Error("Incorrect cost-type count in orig:", nct)
	}

	dir2 := testRegenDirectory(test, "Simple Dir", dir)
	if dir2 != nil {
		nres2 := len(dir2.Resources)
		if nres2 != nres {
			test.Error("Incorrect resource count in regen:", nres2)
		}
		nct2 := len(dir2.CostTypes)
		if nct2 != nct {
			test.Error("Incorrect cost-type count in regen:", nct2)
		}
	}
}

func testAddCostType(ctmap map[string]CostTypeDescription,
				 name, metric, mode, descr string) {
	ct := CostTypeDescription{}
	ct.Metric = metric;
	ct.Mode = mode;
	ct.Description = descr
	ctmap[name] = ct
}

func testRegenDirectory(test *testing.T,
						descr string,
						pmsg *Directory) *Directory {
	altomsg2 := testRegenAltoMsg(test, descr, pmsg)
	if altomsg2 == nil {
		return nil
	} else {
		pmsg2, ok := altomsg2.(*Directory)
		if !ok {
			test.Error(descr, "Regen has wrong type")
			return nil
		} else {
			diff := CmpAltoMsgs(pmsg, pmsg2)
			if diff != "" {
				test.Error(descr, "Regen readback diff:", diff)
			}
			return pmsg2
		}
	}
}

