package altomsgs

import (
	"testing"
	"bytes"
	)

func TestNetworkMapFilter(test *testing.T) {
	pmsg := NewNetworkMapFilter()
	if pmsg.MediaType() != MT_NETWORK_MAP_FILTER {
		test.Error("NewNetworkMapFilter returns wrong media type",
					pmsg.MediaType())
	}
	testRegenAltoMsg(test, "New msg", pmsg)
	testRegenAltoMsg(test, "2 Pids",
				&NetworkMapFilter{
					Pids: []string{"s1", "s2"},
					})
	testRegenAltoMsg(test, "2 Pids & 1 Addrtype",
				&NetworkMapFilter{
					Pids: []string{"s1", "s2"},
					AddrTypes: []string{"ipv4"},
					})
}

func TestCostMapFilter(test *testing.T) {
	pmsg := NewCostMapFilter()
	if pmsg.MediaType() != MT_COST_MAP_FILTER {
		test.Error("NewCostMapFilter returns wrong media type",
					pmsg.MediaType())
	}
	testRegenAltoMsg(test, "New msg", pmsg)
	testRegenAltoMsg(test, "Just costtype",
				&CostMapFilter{
					CostType: CostType{
							Metric: "routingcost",
							Mode: "numerical"},
					})
	testRegenAltoMsg(test, "Everything",
				&CostMapFilter{
					CostType: CostType{
							Metric: "hopcount",
							Mode: "ordinal"},
					Srcs: []string{"s1", "s2"},
					Dsts: []string{"d1", "d2"},
					Constraints: []string{"le 10"},
					})
}

func TestErrorResp(test *testing.T) {
	pmsg := NewErrorResp(ERROR_CODE_MISSING_FIELD)
	if pmsg.MediaType() != MT_ERROR {
		test.Error("NewErrorResp returns wrong media type",
					pmsg.MediaType())
	}
	pmsg.Field = "cost-type"
	testRegenAltoMsg(test, "Missing field", pmsg)
	testRegenAltoMsg(test, "Syntax error",
				&ErrorResp{Code: ERROR_CODE_SYNTAX,
					  SyntaxError: "Line X, col Y: OOPS!",
					  })
	testRegenAltoMsg(test, "Invalid type",
				&ErrorResp{Code: ERROR_CODE_INVALID_FIELD_TYPE,
					  Field: "cost-type.metric",
					  Value: "unknown-cost",
					  })
}

func TestEndpointCostParams(test *testing.T) {
	pmsg := NewEndpointCostParams()
	if pmsg.MediaType() != MT_ENDPOINT_COST_PARAMS {
		test.Error("NewEndpointCostParams returns wrong media type",
					pmsg.MediaType())
	}
	testRegenAltoMsg(test, "New msg", pmsg)
	testRegenAltoMsg(test, "Just costtype",
				&EndpointCostParams{
					CostType: CostType{
							Metric: "routingcost",
							Mode: "numerical"},
					})
	testRegenAltoMsg(test, "Everything",
				&EndpointCostParams{
					CostType: CostType{
							Metric: "hopcount",
							Mode: "ordinal"},
					Srcs: []string{"ipv4:1.2.3.4", "ipv6:1:2:3:4::"},
					Dsts: []string{"ipv4:4.3.2.1", "ipv6:4:3:2:1::"},
					Constraints: []string{"le 10"},
					})
}

func TestEndpointPropParams(test *testing.T) {
	pmsg := NewEndpointPropParams()
	if pmsg.MediaType() != MT_ENDPOINT_PROP_PARAMS {
		test.Error("NewEndpointPropParams returns wrong media type",
					pmsg.MediaType())
	}
	testRegenAltoMsg(test, "New msg", pmsg)
	testRegenAltoMsg(test, "Everything",
				&EndpointPropParams{
					Endpoints: []string{"ipv4:1.2.3.4", "ipv6:1:2:3:4::"},
					Properties: []string{"foo", "bar"},
					})
}

func testRegenAltoMsg(test *testing.T, descr string, pmsg AltoMsg) AltoMsg {
	buff := bytes.Buffer{}
	WriteJson(pmsg, &buff)
	pmsg2, errs := NewAltoMsg(pmsg.MediaType(), &buff, buff.Len())
	if len(errs) > 0 {
		for _, err := range errs {
			test.Error(descr, "NewAltoMsg error:", err)
		}
	} else {
		diff := CmpAltoMsgs(pmsg, pmsg2)
		if diff != "" {
			test.Error(descr, "Readback diff:", diff)
		}
	}
	return pmsg2
}

