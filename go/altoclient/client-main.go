package main

import (
	_ "github.com/wdroome/go/wdrlib"
	"github.com/wdroome/go/altomsgs"
	"fmt"
	"strings"
	"strconv"
	"os"
	"time"
	)

var CmdDescr = []string {
		"ird uri                    ## Fetch a root IRD and prepare to use that server",
		"ird -refresh               ## Re-fetch last root IRD",
		"ird                        ## Print current ALTO server resources",
		"use-netmap [id]            ## Set the network map for netmap & cost commands.",
		"netmap [-addrtype type ...] [-pid pid pid ...] [-id=res-id] [-uri=res-uri]",
		"       [-no-incr] [-tag=[###]]",
		"                           ## Show selected pids (or all pids) in the network map.",
		"                           ## If res-id or res-uri are specified, use that Network Map resource.",
		"                           ## If not, use the current Network Map.",
		"                           ## -tag and -no-incr are used with update-stream commands.",
		"                           ## -no-incr means do not allow incremental updates,",
		"                           ## and -tag=### means the client has the version",
		"                           ## with that tag. -tag=, without a tag value,",
		"                           ## means use the tag for the most recently retrieved",
		"                           ## full network map.",
		"costs [-src pid pid ...] [-dst pid pid ...] [-type=metric/mode]",
		"      [-constraint op value] [-id=res-id] [-uri=res-uri] [-no-incr]",
		"                           ## Show pid costs. If either -src or -dst are present",
		"                           ## use a Filtered Cost Map. Otherwise use a full Cost Map.",
		"                           ## If res-id or res-uri are specified, use that Cost Map resource.",
		"                           ## If not, pick the appropriate Cost Map resource.",
		"                           ## -no-incr is used with update-stream commands,.",
		"                           ## and means do not allow incremental updates.",
		"end-costs [-src addr addr ...] [-dst addr addr ...] [-type=metric/mode]",
		"          [-constraint op value] [-id=res-id] [-uri=res-uri] [-no-incr]",
		"                           ## Show endpoint costs.",
		"                           ## If res-id or res-uri are specified, use that resource.",
		"                           ## If not, pick the appropriate Endpoint Cost resource.",
		"                           ## -no-incr is used with update-stream commands,.",
		"                           ## and means do not allow incremental updates.",
		"props [-addr addr addr ...] [-prop prop prop ...]",
		"      [-id=res-id] [-uri=res-uri] [-no-incr]",
		"                           ## Show endpoint properties.",
		"                           ## If res-id or res-uri are specified, use that resource.",
		"                           ## If not, pick the appropriate Endpoint Prop resource.",
		"                           ## -no-incr is used with update-stream commands,.",
		"                           ## and means do not allow incremental updates.",
		"show [netmaps] [filtered-netmaps] [costmaps] [filtered-costmaps]",
		"     [metric=cost-metric] [end-costs] [end-props]",
		"                           ## Show the IRD information for all resources",
		"                           ## of the indicated types.",
		"find-pids addr addr ...    ## Show PIDs for addresses.",
		"                           ## You must fetch a full Network Map first.",
		"find-cidrs pid pid ...     ## Show CIDRs for pids.",
		"                           ## You must fetch a full Network Map first.",
		"find-costs -src pid pid ... -dst pid pid ...",
		"                           ## Show costs for pids in the last full Cost Map.",
		"timeout [duration]         ## Set or show the request timeout.",
		"                           ## The duration can be in any format",
		"                           ## accepted by time.ParseDuration,",
		"                           ## such as 200s, 200000ms, etc.",
		"proxy [uri]                ## Set or show an HTTP proxy",
		"skip-verify [true|false]   ## Set 'promiscous' mode. If true, accept all https",
		"                           ## server credentials, even if they are self-signed",
		"                           ## or the host name doesn't match.",
		"help                       ## Print a summary of all commands",
		"help word word ...         ## Print the help items containing those words",
		"quit                       ## The obvious",
	}

/*
	Java client commands which have not been implemented in the go client:
	"validate-msgs [true|false] ## Whether to validate server responses.",
	"print-msgs [true|false]    ## Whether to print the JSON response messages.",
	"read-cmds filename         ## Read commands from a file.",
*/

const (
	URI_ARG = "-uri"
	ID_ARG = "-id"
	TAG_ARG = "-tag"
	NO_INCR_ARG = "-no-incr"
	PID_ARG = "-pid"
	ADDR_TYPE_ARG = "-addrtype"
	SRC_ARG = "-src"
	DST_ARG = "-dst"
	TYPE_ARG = "-type"
	ADDR_ARG = "-addr"
	PROP_ARG = "-prop"
	CONSTRAINT_ARG = "-constraint"
	)
	
var altoConn *altomsgs.AltoConn
var lastFullNetMap *altomsgs.NetworkMap
var lastFullCostMap *altomsgs.CostMap

func main() {
	rdr := NewCmdReader(nil)
	altoConn = altomsgs.NewAltoConn()
	
	if len(os.Args) == 2 &&
			(strings.HasPrefix(os.Args[1], "http:") ||
			 strings.HasPrefix(os.Args[1], "https:")) {
		IRDCmd([]string{os.Args[1]})
	}	

	for {
		cmd := rdr.NextCmd()
		if cmd == nil {
			break
		}
		if len(cmd) == 0 {
			continue
		}
		switch cmd[0] {
		case "quit":
			return
		case "q":
			return
		case "help":
			HelpCmd(cmd[1:])
		case "?":
			HelpCmd(cmd[1:])
		case "ird":
			IRDCmd(cmd[1:])
		case "use-netmap":
			UseNetmapCmd(cmd[1:])
		case "netmap":
			NetmapCmd(cmd[1:])
		case "costs":
			CostsCmd(cmd[1:])
		case "end-costs":
			EndCostsCmd(cmd[1:])
		case "props":
			PropsCmd(cmd[1:])
		case "show":
			ShowCmd(cmd[1:])
		case "find-pids":
			FindPidsCmd(cmd[1:])
		case "find-cidrs":
			FindCidrsCmd(cmd[1:])
		case "find-costs":
			FindCostsCmd(cmd[1:])
		case "timeout":
			TimeoutCmd(cmd[1:])
		case "proxy":
			ProxyCmd(cmd[1:])
		case "skip-verify":
			SkipVerifyCmd(cmd[1:])
		default:
			fmt.Println("Unknown command", cmd[0])
		}
	}
}

func IRDCmd(args []string) {
	if len(args) >= 1 {
		uri := ""
		if args[0] == "-refresh" {
			if altoConn.ResourceSet != nil {
				uri = altoConn.ResourceSet.URI
			}
		} else {
			uri = args[0]
		}
		if uri == "" {
			fmt.Println("No URI specified.")
			return
		}
		respTime, errs := altoConn.LoadRootDir(uri)
		if len(errs) > 0 {
			printErrs(errs)
		}
		if !altoConn.HaveResources {
			fmt.Println("Load failed: no resources")
		} else {
			fmt.Printf("Loaded %s in %s\n", uri, respTime.String())
			fmt.Printf("  %d Resources  Default Netmap: %s\n",
						len(altoConn.ResourceSet.Resources),
						altoConn.NetworkMapId)
		}
	} else if !altoConn.HaveResources {
		fmt.Println("No IRD")
	} else {
		fmt.Println("Num resources: %d netmap: %s\n",
					len(altoConn.ResourceSet.Resources),
					altoConn.NetworkMapId)
		altoConn.ResourceSet.Print(os.Stdout)
	}
}

func UseNetmapCmd(args []string) {
	if !ConnExists() {
		return
	}
	if len(args) >= 1 {
		altoConn.NetworkMapId = args[0]
	}
	fmt.Println("Network Map Id: \"" + altoConn.NetworkMapId + "\"")
}

func HelpCmd(args []string) {
	lineSet := []string{}
	prtLineSet := len(args) == 0
	for _, line := range CmdDescr {
		if !strings.HasPrefix(line, " ") {
			if prtLineSet {
				for _, s := range lineSet {
					fmt.Println(s)
				}
			}
			lineSet = []string{}
			prtLineSet = len(args) == 0
		}
		lineSet = append(lineSet, line)
		for _, s := range args {
			if strings.Contains(strings.ToLower(line), strings.ToLower(s)) {
				prtLineSet = true
			}
		}
	}
	if prtLineSet {
		for _, s := range lineSet {
			fmt.Println(s)
		}
	}
}

func ProxyCmd(args []string) {
	if len(args) == 0 {
		fmt.Print("Current proxy: ")
		if altoConn.Proxy == nil {
			fmt.Println("none")
		} else {
			fmt.Println(altoConn.Proxy.String())
		}
	} else if len(args) == 1 {
		err := altoConn.SetProxy(args[0])
		if err != nil {
			fmt.Println("Error:", err)
		}
	} else {
		fmt.Println("Usage: proxy [proxy-uri]")
	}
}

func TimeoutCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Current timeout: " + altoConn.Timeout().String())
	} else if len(args) == 1 {
		timeout, err := time.ParseDuration(args[0])
		if err != nil {
			fmt.Println("Invalid timeout:", err)
		}
		altoConn.SetTimeout(timeout)
	} else {
		fmt.Println("Usage: timeout [timeout]")
	}
}

func SkipVerifyCmd(args []string) {
	if len(args) == 0 {
		fmt.Println("Skip-verify mode:", altoConn.SkipVerify())
	} else if len(args) == 1 {
		skip, err := strconv.ParseBool(args[0])
		if err != nil {
			fmt.Println("Invalid boolean:", err)
		}
		altoConn.SetSkipVerify(skip)
	} else {
		fmt.Println("Usage: skip-verify [true|false]")
	}
}

// DoReq() sends an ALTO request and returns the server's response.
// "uri" is the URI of an ALTO resource, and "accept" has the
// media type(s) the client is willing to accept
// (the function adds MT_ERROR to the accept list,
// if not already present).
// If "req" is nil, send a GET request.
// If "req" is not nil, send a POST request with that message as the body.
func DoReq(uri string, accept []string, req altomsgs.AltoMsg) *altomsgs.ServerResp {
	fmt.Println("  Sending to " + uri + ":")
	servResp := altoConn.SendReq(uri, accept, req)
	if servResp.HaveResponse {
		fmt.Printf("  HTTP Status: %s  Len: %d  Time: %s\n",
					servResp.Status, servResp.ContentLength,
					servResp.RespTime.String())
	}
	if len(servResp.Errors) > 0 {
		fmt.Println("Errors:")
		for _, err := range servResp.Errors {
			fmt.Println("  " + err.Error())
		}
	}
	if servResp.ErrorResp != nil {
		fmt.Println("  ALTO Error, Code: " + servResp.ErrorResp.Code)
		if servResp.ErrorResp.SyntaxError != "" {
			fmt.Println("      SyntaxError: \"" + servResp.ErrorResp.SyntaxError + "\"")
		}
		if servResp.ErrorResp.Field != "" {
			fmt.Println("      Field: \"" + servResp.ErrorResp.Field + "\"")
		}
		if servResp.ErrorResp.Value != "" {
			fmt.Println("      Value: \"" + servResp.ErrorResp.Value + "\"")
		}
	} else if servResp.OkResp != nil {
		fmt.Println("  ALTO Response Media-Type: " + servResp.OkResp.MediaType())
		altomsgs.PrintAltoMsg(servResp.OkResp, os.Stdout)
	}
	return servResp
}
	
// ConnExists() returns true iff we have downloaded the IRD from the ALTO server.
// If not, the function prints an error message and returns false.
func ConnExists() bool {
	if altoConn == nil || !altoConn.HaveResources {
		fmt.Println("There is no connection to the ALTO server.")
		return false
	} else {
		return true
	}
}

// NetMapExists() returns true iff a network map id has been specifed.
// If not, the function prints an error message and returns false.
func NetMapExists() bool {
	if altoConn == nil || altoConn.NetworkMapId == "" {
		fmt.Println("Please use \"use-netmap\" to specify a network map.")
		return false
	} else {
		return true
	}
}

func printErrs(errs []error) {
	if len(errs) == 1 {
		fmt.Println("ERROR:", errs[0])
	} else {
		fmt.Println(len(errs), "Errors:")
		for _, err := range errs {
			fmt.Println("  ", err)
		}
	}
}
