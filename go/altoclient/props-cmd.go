package main

import (
	_ "github.com/wdroome/go/wdrlib"
	"github.com/wdroome/go/altomsgs"
	"fmt"
	"strings"
	)

var PropsCmd_LegalArgs = LegalArgs{
				Names: []string{URI_ARG, ID_ARG},
				Lists: []string{ADDR_ARG, PROP_ARG},
				Flags: []string{NO_INCR_ARG},
				}

func PropsCmd(args []string) {
	if !ConnExists() {
		return
	}
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &PropsCmd_LegalArgs)
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
	addrs := parsedArgs.Lists[ADDR_ARG]
	props := parsedArgs.Lists[PROP_ARG]
	uri := parsedArgs.Names[URI_ARG]
	
	if uri == "" {
		res := altoConn.ResourceSet.FindEndpointProp(props)
		if res == nil {
			fmt.Println("The server does not provide an endpoint property " +
						"resource for " + strings.Join(props, " "))
			return
		}
		uri = res.URI.String()
	}
	reqMsg := &altomsgs.EndpointPropParams{Endpoints: addrs,
										   Properties: props}
	servResp := DoReq(uri, []string{altomsgs.MT_ENDPOINT_PROP}, reqMsg)
	if servResp.OkResp != nil {
		switch v := servResp.OkResp.(type) {
		case *altomsgs.EndpointProp:
			// Okay
			_ = v
		default:
			fmt.Println("ERROR: Wrong response type " +
							servResp.OkResp.MediaType())
		}
	}
}
