package main

import (
	_ "github.com/wdroome/go/wdrlib"
	"github.com/wdroome/go/altomsgs"
	"fmt"
	_ "net"
	)

var NetmapCmd_LegalArgs = LegalArgs{
				Names: []string{URI_ARG, ID_ARG, TAG_ARG},
				Lists: []string{PID_ARG, ADDR_TYPE_ARG},
				Flags: []string{NO_INCR_ARG},
				}

func NetmapCmd(args []string) {
	if !ConnExists() {
		return
	}
	isFullNetMap := false
	parsedArgs := ParsedArgs{}
	parsedArgs.Parse(args, &NetmapCmd_LegalArgs)
	if !parsedArgs.URIFromId() {
		return
	}
	addrTypes := parsedArgs.Lists[ADDR_TYPE_ARG]
	pids := parsedArgs.Lists[PID_ARG]
	for _, x := range parsedArgs.Lists[""] {
		if x == altomsgs.IPV4_ADDR_TYPE || x == altomsgs.IPV6_ADDR_TYPE {
			addrTypes = append(addrTypes, x)
		} else {
			pids = append(pids, x)
		}
	}
	uri := parsedArgs.Names[URI_ARG]
	id := parsedArgs.Names[ID_ARG]
	var reqMsg altomsgs.AltoMsg = nil

	if addrTypes == nil && pids == nil {
		// Full netmap
		isFullNetMap = true
		if uri == "" {
			if !NetMapExists() {
				return
			}
			id = altoConn.NetworkMapId
			res, ok := altoConn.ResourceSet.Resources[id]
			if !ok {
				fmt.Println("No resource with ID \"" + id + "\"")
				return
			}
			uri = res.URI.String()
		}
	} else {
		// Filtered netmap
		reqMsg = &altomsgs.NetworkMapFilter{AddrTypes: addrTypes, Pids: pids}
		if uri == "" {
			if !NetMapExists() {
				return
			}
			res := altoConn.ResourceSet.FindFilteredNetworkMap(altoConn.NetworkMapId)
			if res == nil {
				fmt.Println("The server does not provide a filtered " +
							"network map resource for \"" +
							altoConn.NetworkMapId + "\"")
				return
			}
			id = res.Id
			uri = res.URI.String()
		}
	}
	servResp := DoReq(uri, []string{altomsgs.MT_NETWORK_MAP}, reqMsg)
	if servResp.OkResp != nil {
		switch v := servResp.OkResp.(type) {
		case *altomsgs.NetworkMap:
			if isFullNetMap {
				lastFullNetMap = v
			}
		default:
			fmt.Println("ERROR: Wrong response type " +
							servResp.OkResp.MediaType())
		}
	}
}

func FindPidsCmd(args []string) {
	if lastFullNetMap == nil {
		fmt.Println("You must fetch a full network map first")
		return
	}
	for _, addr := range args {
		fmt.Print(addr + ": ")
		ip, err := altomsgs.ParseTypedAddr(addr)
		if err != nil {
			fmt.Println(err)
		} else {
			pid, cidr, ok := lastFullNetMap.IP2Pid(ip)
			if !ok {
				fmt.Println("No PID")
			} else {
				fmt.Println("PID:", pid, " CIDR:", cidr.String())
			}
		}
	}
}

func FindCidrsCmd(args []string) {
	if lastFullNetMap == nil {
		fmt.Println("You must fetch a full network map first")
		return
	}
	for _, pid := range args {
		fmt.Println(pid + ":")
		addrTypes, ok := lastFullNetMap.PidAddrs(pid)
		if !ok {
			fmt.Println("  No PID")
		} else {
			for addrType, cidrs := range addrTypes {
				fmt.Print("  " + addrType + ":")
				for _, cidr := range cidrs {
					fmt.Print(" " + cidr)
				}
				fmt.Println()
			}
		}
	}
}

