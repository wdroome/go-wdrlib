package main

import (
	"github.com/wdroome/go/wdrlib"
	"github.com/wdroome/go/altomsgs"
	"fmt"
	"os"
	)

const (
	NETMAPS_ARG = "netmaps"
	FILTERED_NETMAPS_ARG = "filtered-netmaps"
	COSTMAPS_ARG = "costmaps"
	FILTERED_COSTMAPS_ARG = "filtered-costmaps"
	METRIC_ARG = "metric"
	END_COSTS_ARG = "end-costs"
	END_PROPS_ARG = "end-props"
)

var ShowCmd_LegalArgs = LegalArgs{
				Names: []string{METRIC_ARG},
				Flags: []string{NETMAPS_ARG, FILTERED_NETMAPS_ARG,
								COSTMAPS_ARG, FILTERED_COSTMAPS_ARG,
								END_COSTS_ARG, END_PROPS_ARG,
								},
				}

func ShowCmd(args []string) {
	if !ConnExists() {
		return
	}
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &ShowCmd_LegalArgs)
	indent := "  "
	
	if wdrlib.StrListContains(parsedArgs.Flags, NETMAPS_ARG) {
		fmt.Println("Netmap Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_NETWORK_MAP &&
						res.Accepts == "" {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	if wdrlib.StrListContains(parsedArgs.Flags, FILTERED_NETMAPS_ARG) {
		fmt.Println("Filtered Netmap Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_NETWORK_MAP &&
						res.Accepts == altomsgs.MT_NETWORK_MAP_FILTER {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	if wdrlib.StrListContains(parsedArgs.Flags, COSTMAPS_ARG) {
		fmt.Println("Costmap Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_COST_MAP &&
						res.Accepts == "" {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	if wdrlib.StrListContains(parsedArgs.Flags, FILTERED_COSTMAPS_ARG) {
		fmt.Println("Filtered Netmap Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_COST_MAP &&
						res.Accepts == altomsgs.MT_COST_MAP_FILTER {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	if wdrlib.StrListContains(parsedArgs.Flags, END_COSTS_ARG) {
		fmt.Println("Endpoint Cost Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_ENDPOINT_COST &&
						res.Accepts == altomsgs.MT_ENDPOINT_COST_PARAMS {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	if wdrlib.StrListContains(parsedArgs.Flags, END_PROPS_ARG) {
		fmt.Println("Endpoint Property Resources:")
		for _, res := range altoConn.ResourceSet.Resources {
			if res.MediaType == altomsgs.MT_ENDPOINT_PROP &&
						res.Accepts == altomsgs.MT_ENDPOINT_PROP_PARAMS {
				res.Print(os.Stdout, indent)
			}
		}
	}
	
	metric, ok := parsedArgs.Names[METRIC_ARG]
	if ok {
		fmt.Println("Costmaps for metric \"" + metric + "\":")
		for _, res := range altoConn.ResourceSet.Resources {
			for _, ct := range res.CostTypes {
				if ct.Metric == metric {
					res.Print(os.Stdout, indent)
					break
				}
			}
		}
	}

	if len(parsedArgs.Lists[""]) > 0 {
		fmt.Print("Unknown arguments:")
		for _, x := range parsedArgs.Lists[""] {
			fmt.Print(" " + x)
		}
		fmt.Println()
		return
	}
}
				
