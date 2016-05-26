package altomsgs

import (
	"testing"
	_ "fmt"
)

func TestCostTypes(test *testing.T) {
	cts := []CostType{
				CostType{"routingcost", "numerical"},
				CostType{"routingcost", "ordinal"},
				CostType{"hopcount", "numerical"},
			}
	for _, ct := range cts {
		if !CostTypeListContains(cts, ct) {
			test.Error("CostTypeListContains: did not find", ct)
		}
	}
	xct := CostType{"hopcount", "ordinal"}
	if CostTypeListContains(cts, xct) {
		test.Error("CostTypeListContains: found", xct)
	}
	ctsEq := []CostType{
				CostType{"routingcost", "numerical"},
				CostType{"routingcost", "ordinal"},
				CostType{"hopcount", "numerical"},
			}
	ctsNe1 := []CostType{
				CostType{"routingcost", "numerical"},
				CostType{"hopcount", "numerical"},
				CostType{"routingcost", "ordinal"},
			}
	ctsNe2 := []CostType{
				CostType{"routingcost", "numerical"},
				CostType{"routingcost", "ordinal"},
			}
			
	if !CostTypeListEqual(cts, ctsEq) {
		test.Error("CostTypeListEqual: says cts != ctsEq1")
	}
	if CostTypeListEqual(cts, nil) {
		test.Error("CostTypeListEqual: says cts == nil")
	}
	if CostTypeListEqual(nil, cts) {
		test.Error("CostTypeListEqual: says nil == cts")
	}
	if CostTypeListEqual(cts, ctsNe1) {
		test.Error("CostTypeListEqual: says cts == ctsNe1")
	}
	if CostTypeListEqual(cts, ctsNe2) {
		test.Error("CostTypeListEqual: says cts == ctsNe2")
	}
			
	if !CostTypeSetEqual(cts, ctsEq) {
		test.Error("CostTypeSetEqual: says cts != ctsEq1")
	}
	if CostTypeSetEqual(cts, nil) {
		test.Error("CostTypeSetEqual: says cts == nil")
	}
	if CostTypeSetEqual(nil, cts) {
		test.Error("CostTypeSetEqual: says nil == cts")
	}
	if !CostTypeSetEqual(cts, ctsNe1) {
		test.Error("CostTypeSetEqual: says cts != ctsNe1")
	}
	if CostTypeSetEqual(cts, ctsNe2) {
		test.Error("CostTypeSetEqual: says cts == ctsNe2")
	}
	if !CostTypeSetEqual(nil, nil) {
		test.Error("CostTypeSetEqual: says nil != nil")
	}
}
