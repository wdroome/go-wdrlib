package wdrlib

import (
	"testing"
	_ "fmt"
)

func TestStrUtilFuncs(test *testing.T) {
	ss := []string{"abc", "def", "deg"}
	for _, s := range ss {
		if !StrListContains(ss, s) {
			test.Error("StrListContains: did not find", s)
		}
	}
	if StrListContains(ss, "dex") {
		test.Error("StrListContains: found", "dex")
	}
	
	ss1 := []string{"abc", "def", "deg"}
	ss2 := []string{"abc", "deg", "def"}
	ss3 := []string{"abc", "def"}
	
	if !StrListEqual(ss, ss1) {
		test.Error("StrListEqual: says ss != ss1")
	}
	if StrListEqual(ss, nil) {
		test.Error("StrListEqual: says ss == nil")
	}
	if StrListEqual(nil, ss) {
		test.Error("StrListEqual: says nil == ss")
	}
	if StrListEqual(ss, ss2) {
		test.Error("StrListEqual: says ss == ss2")
	}
	if StrListEqual(ss, ss3) {
		test.Error("StrListEqual: says ss == ss3")
	}
	
	if !StrSetEqual(ss, ss1) {
		test.Error("StrSetEqual: says ss != ss1")
	}
	if StrSetEqual(ss, nil) {
		test.Error("StrSetEqual: says ss == nil")
	}
	if StrSetEqual(nil, ss) {
		test.Error("StrSetEqual: says nil == ss")
	}
	if !StrSetEqual(ss, ss2) {
		test.Error("StrSetEqual: says ss != ss2")
	}
	if StrSetEqual(ss, ss3) {
		test.Error("StrSetEqual: says ss == ss3")
	}
	if !StrSetEqual(nil, nil) {
		test.Error("StrSetEqual: says nil != nil")
	}
}

func TestIFaceArrToStrs(test *testing.T) {
	var sarr []string
	
	sarr = IfaceArrToStrs([]interface{}{"a", 1, 2.1, true})
	if !StrListEqual(sarr, []string{"a", "1", "2.1", "true"}) {
		test.Error("IfaceArrToStrs: returns", sarr,
					"for []interface{}{\"a\", 1, 2.1, true}")
	}
	
	sarr = IfaceArrToStrs([]interface{}{"a", "b"})
	if !StrListEqual(sarr, []string{"a", "b"}) {
		test.Error("IfaceArrToStrs: returns", sarr,
					"for []interface{}{\"a\", \"b\"}")
	}
}
