package main

import (
	"github.com/wdroome/go/wdrlib"
	_ "github.com/wdroome/go/altomsgs"
	"fmt"
	"strings"
	)

// ParsedArgs represents comamnd line arguments.
// Names are the  name=value  arguments.
// Lists are the  name value value ...  arguments.
// Flags are the  name  arguments.
// For Names & Lists, the name is the key, and the value is the value.
// For Flags, the name is in the string slice iff it was a parameter.
// Unknown arguments are in Lists[""].
type ParsedArgs struct {
	Names map[string]string
	Lists map[string][]string
	Flags []string
}

// LegalArgs describes the acceptable arguments & their types.
type LegalArgs struct {
	Names []string
	Lists []string
	Flags []string
}

// Parse() parses a list of arguments and saves them in the appropriate
// ParsedArgs members.
func (this *ParsedArgs) Parse(args []string, legal *LegalArgs) {
	if this.Names == nil {
		this.Names = map[string]string{}
	}
	if this.Lists == nil {
		this.Lists = map[string][]string{}
	}
	curList := ""
	for _, arg := range args {
		nameVal := strings.SplitN(arg, "=", 2)
		used := false
		switch len(nameVal) {
		case 1:
			if wdrlib.StrListContains(legal.Flags, arg) {
				this.Flags = append(this.Flags, arg)
				curList = ""
				used = true
			} else if wdrlib.StrListContains(legal.Lists, arg) {
				curList = arg
				if this.Lists[curList] == nil {
					this.Lists[curList] = []string{}
				}
				used = true
			}
		case 2:
			if wdrlib.StrListContains(legal.Names, nameVal[0]) {
				this.Names[nameVal[0]] = nameVal[1]
				curList = ""
				used = true
			}
		}
		if !used {
			this.Lists[curList] = append(this.Lists[curList], arg)
		}
	}
}

// URIFromId() sets the URI argument to the URI of the
// resource with ID, if ID is specified.
// It returns true if URI has been set from ID,
// or if no ID arguent was specified.
// It returns false iff ID was specified, but no resource has that ID. 
func (this *ParsedArgs) URIFromId() bool {
	id, ok := this.Names[ID_ARG]
	if !ok {
		return true
	}
	res, ok := altoConn.ResourceSet.Resources[id]
	if !ok {
		fmt.Println("No resource with ID \"" + id + "\"")
		return false
	}
	this.Names[URI_ARG] = res.URI.String()
	return true
}
