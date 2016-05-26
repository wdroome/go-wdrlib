package main

import (
	_ "github.com/wdroome/go/wdrlib"
	"github.com/wdroome/go/altomsgs"
	"fmt"
	"strings"
	)

var CostsCmd_LegalArgs = LegalArgs{
				Names: []string{TYPE_ARG, URI_ARG, ID_ARG},
				Lists: []string{SRC_ARG, DST_ARG, CONSTRAINT_ARG},
				Flags: []string{NO_INCR_ARG},
				}

func CostsCmd(args []string) {
	if !ConnExists() {
		return
	}
	if !NetMapExists() {
		return
	}
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &CostsCmd_LegalArgs)
	if !parsedArgs.URIFromId() {
		return
	}
	if len(parsedArgs.Lists[""]) > 0 {
		fmt.Print("Unknown arguments:")
		for _, x := range parsedArgs.Lists[""] {
			fmt.Print(" " + x)
		}
		fmt.Println()
		return
	}
	srcs := parsedArgs.Lists[SRC_ARG]
	dsts := parsedArgs.Lists[DST_ARG]
	uri := parsedArgs.Names[URI_ARG]
	costType := parsedArgs.parseTypeArg()
	constraints := parsedArgs.parseConstraintArg()
	
	var reqMsg altomsgs.AltoMsg = nil
	isFullCostMap := false
	if srcs == nil && dsts == nil && constraints == nil {
		// Full costmap
		isFullCostMap = true
		if uri == "" {
			res := altoConn.ResourceSet.FindCostMap(
									altoConn.NetworkMapId,
									costType)
			if res == nil {
				fmt.Println("The server does not provide a " +
							"cost map resource for CostType " + costType.String())
				return
			}
			uri = res.URI.String()
		}
	} else {
		// Filtered costmap
		if uri == "" {
			res := altoConn.ResourceSet.FindFilteredCostMap(
									altoConn.NetworkMapId,
									costType,
									constraints != nil)
			if res == nil {
				fmt.Println("The server does not provide a filtered " +
							"cost map resource for CostType " + costType.String())
				return
			}
			uri = res.URI.String()
		}
		reqMsg = &altomsgs.CostMapFilter{Srcs: srcs, Dsts: dsts,
								CostType: costType, Constraints: constraints}
	}
	servResp := DoReq(uri, []string{altomsgs.MT_COST_MAP}, reqMsg)
	if servResp.OkResp != nil {
		switch v := servResp.OkResp.(type) {
		case *altomsgs.CostMap:
			if isFullCostMap {
				lastFullCostMap = v
			}
		default:
			fmt.Println("ERROR: Wrong response type " +
							servResp.OkResp.MediaType())
		}
	}
}

var EndCostsCmd_LegalArgs = LegalArgs{
				Names: []string{TYPE_ARG, URI_ARG, ID_ARG},
				Lists: []string{SRC_ARG, DST_ARG, CONSTRAINT_ARG},
				Flags: []string{NO_INCR_ARG},
				}

func EndCostsCmd(args []string) {
	if !ConnExists() {
		return
	}
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &EndCostsCmd_LegalArgs)
	if !parsedArgs.URIFromId() {
		return
	}
	if len(parsedArgs.Lists[""]) > 0 {
		fmt.Print("Unknown arguments:")
		for _, x := range parsedArgs.Lists[""] {
			fmt.Print(" " + x)
		}
		fmt.Println()
		return
	}
	srcs := parsedArgs.Lists[SRC_ARG]
	dsts := parsedArgs.Lists[DST_ARG]
	uri := parsedArgs.Names[URI_ARG]
	costType := parsedArgs.parseTypeArg()
	constraints := parsedArgs.parseConstraintArg()
	
	if uri == "" {
		res := altoConn.ResourceSet.FindEndpointCost(costType, constraints != nil)
		if res == nil {
			fmt.Println("The server does not provide a endpoint cost " +
						"resource for CostType " + costType.String())
			return
		}
		uri = res.URI.String()
	}
	reqMsg := &altomsgs.EndpointCostParams{Srcs: srcs, Dsts: dsts,
							CostType: costType, Constraints: constraints}
	servResp := DoReq(uri, []string{altomsgs.MT_ENDPOINT_COST}, reqMsg)
	if servResp.OkResp != nil {
		switch v := servResp.OkResp.(type) {
		case *altomsgs.EndpointCost:
			// Okay
			_ = v
		default:
			fmt.Println("ERROR: Wrong response type " +
							servResp.OkResp.MediaType())
		}
	}
}

var FindCostsCmd_LegalArgs = LegalArgs{
				Lists: []string{SRC_ARG, DST_ARG},
				}

func FindCostsCmd(args []string) {
	if lastFullCostMap == nil {
		fmt.Println("You must fetch a full cost map first")
		return
	}
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &FindCostsCmd_LegalArgs)
	if len(parsedArgs.Lists[""]) > 0 {
		fmt.Print("Unknown arguments:")
		for _, x := range parsedArgs.Lists[""] {
			fmt.Print(" " + x)
		}
		fmt.Println()
		return
	}
	srcs := parsedArgs.Lists[SRC_ARG]
	dsts := parsedArgs.Lists[DST_ARG]
	if len(srcs) <= 0 {
		srcs = lastFullCostMap.AllSrcs()
	}
	if len(dsts) <= 0 {
		dsts = lastFullCostMap.AllDsts()
	}
	
	fmt.Println("CostType: " + lastFullCostMap.CostType().String())
	for _, src := range srcs {
		n := 0
		fmt.Print(src + ":")
		for _, dst := range dsts {
			cost, ok := lastFullCostMap.GetCost(src, dst)
			if ok {
				if n >= 5 {
					fmt.Println()
					n = 0
				}
				fmt.Printf("  %s: %g", dst, cost)
				n++
			}
		}
		fmt.Println()
	}
}

func (this *ParsedArgs) parseTypeArg() altomsgs.CostType {
	val, ok := this.Names[TYPE_ARG]
	if !ok {
		return altomsgs.CostType{Metric: altomsgs.CT_ROUTINGCOST,
								 Mode: altomsgs.CT_NUMERICAL}
	}
	x := strings.SplitN(val, "/", 2)
	switch len(x) {
	case 2:
		return altomsgs.CostType{Metric: x[0], Mode: x[1]}
	case 1:
		return altomsgs.CostType{Metric: x[0], Mode: altomsgs.CT_NUMERICAL}
	default:
		return altomsgs.CostType{}
	}
}

func (this *ParsedArgs) parseConstraintArg() []string {
	val, ok := this.Lists[CONSTRAINT_ARG]
	if !ok {
		return nil
	}
	n := len(val)
	var list []string = nil
	for i := 0; i+1 < n; i += 2 {
		list = append(list, val[i] + " " + val[i+1])
	}
	return list
}
