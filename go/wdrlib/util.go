package wdrlib

import (
	"fmt"
	"sort"
	"encoding/json"
	"io"
	)

// PrintMapTree() prints an object tree created by json.Unmarshal()
// in a consistent, repeatable fashion.
func PrintMapTree(x interface{}, w io.Writer, indent string) {
	indentIncr := "   "
	switch xv := x.(type) {
	case nil:
		fmt.Fprintf(w, "%snil\n", indent);
	case string:
		fmt.Fprintf(w, "%s%q\n", indent, xv);
	case int:
		fmt.Fprintf(w, "%s%d\n", indent, xv)
	case float32:
		fmt.Fprintf(w, "%s%f\n", indent, xv)
	case float64:
		fmt.Fprintf(w, "%s%f\n", indent, xv)
	case bool:
		fmt.Fprintf(w, "%s%t\n", indent, xv)
	case map[string]interface{}:
		keys := make([]string, len(xv))
		i := 0
		for k := range xv {
			keys[i] = k
			i++
		}
		sort.Strings(keys)
		for _, k := range keys {
			switch xxv := xv[k].(type) {
			case nil:
				fmt.Fprintf(w, "%s%s: nil\n", indent, k);
			case string:
				fmt.Fprintf(w, "%s%s: %q\n", indent, k, xxv);
			case int:
				fmt.Fprintf(w, "%s%s: %d\n", indent, k, xxv)
			case float32:
				fmt.Fprintf(w, "%s%s: %f\n", indent, k, xxv)
			case float64:
				fmt.Fprintf(w, "%s%s: %f\n", indent, k, xxv)
			case bool:
				fmt.Fprintf(w, "%s%s: %t\n", indent, k, xxv)
			default:
				fmt.Fprintf(w, "%s%s:\n", indent, k)
				PrintMapTree(xv[k], w, indent + indentIncr)
			}
		}
	case []interface{}:
		for i, u := range xv {
			switch uv := u.(type) {
			case nil:
				fmt.Fprintf(w, "%s[%d]: nil\n", indent, i);
			case string:
				fmt.Fprintf(w, "%s[%d]: %q\n", indent, i, uv);
			case int:
				fmt.Fprintf(w, "%s[%d]: %d\n", indent, i, uv)
			case float32:
				fmt.Fprintf(w, "%s[%d]: %f\n", indent, i, uv)
			case float64:
				fmt.Fprintf(w, "%s[%d]: %f\n", indent, i, uv)
			case bool:
				fmt.Fprintf(w, "%s[%d]: %t\n", indent, i, uv)
			default:
				fmt.Fprintf(w, "%s[%d]:\n", indent, i)
				PrintMapTree(u, w, indent + indentIncr)
			}
		}
	default:
		fmt.Fprintf(w, "%s(Unknown type)\n", indent)
	}		
}

// PrintJson() prints a JSON string in a consistent, repeatable fashion.
// It returns an error if jsonBytes is not valid JSON.
func PrintJson(jsonBytes []byte, w io.Writer) error {
	var jm interface{}
	err := json.Unmarshal(jsonBytes, &jm)
	if err != nil {
		fmt.Fprintln(w, "Unmarshal failed:", err)
		return err
	} else {
		PrintMapTree(jm, w, "")
		return nil
	}
}

// IfaceArrToStrs() returns a []string with the string values
// of the elements in a general array.
func IfaceArrToStrs(xarr []interface{}) []string {
	if xarr == nil {
		return []string{}
	}
	strs := make([]string, 0, len(xarr))
	for _, v := range xarr {
		switch vv := v.(type) {
		case string:
			strs = append(strs, vv)
		case float64:
			strs = append(strs, fmt.Sprintf("%g", vv))
		case int:
			strs = append(strs, fmt.Sprintf("%d", vv))
		case bool:
			strs = append(strs, fmt.Sprintf("%t", vv))
		}
	}
	return strs
}

// GetStringMember() returns m[k] if that is a string, or "" otherwise.
func GetStringMember(m map[string]interface{}, k string) string {
	v, ok := m[k].(string)
	if ok {
		return v
	} else {
		return ""
	}
}
	
// GetStringArray() returns m[k] if that is a []string, or def otherwise.
func GetStringArray(m map[string]interface{}, k string, def []string) []string {
	v, ok := m[k].([]interface{})
	if ok {
		return IfaceArrToStrs(v)
	} else {
		return nil
	}
}

// GetBoolMember() returns m[k] if that is a boolean, or def otherwise.
func GetBoolMember(m map[string]interface{}, k string, def bool) bool {
	v, ok := m[k].(bool)
	if ok {
		return v
	} else {
		return def
	}
}

// GetFloat64Member() returns m[k] if that is a float64, or def otherwise.
func GetFloat64Member(m map[string]interface{}, k string, def float64) float64 {
	v, ok := m[k].(float64)
	if ok {
		return v
	} else {
		return def
	}
}

// AppendErrors() appends each error in "add" to the slice "cur",
// and returns the (possibly resized) slice.
func AppendErrors(cur []error, add []error) (errs []error) {
	if cur == nil {
		cur = []error{}
	}
	errs = cur
	for _, e := range add {
		errs = append(errs, e)
	}
	return
}

// StrListContains(list,s) returns true iff s equals a string in list.
// Return false if list is nil.
// If the standard library has this function, I missed it.
func StrListContains(list []string, s string) bool {
	for _, elem := range list {
		if s == elem {
			return true
		}
	}
	return false
}

// StrListContainsAll(list,strs) returns true iff every string
// in strs equals a string in list.
// If strs is nil or 0-length, always return true.
// Otherwise, return false if list is nil.
// If the standard library has this function, I missed it.
func StrListContainsAll(list []string, strs []string) bool {
	for _, s := range strs {
		if !StrListContains(list, s) {
			return false
		}
	}
	return true
}

// StrListEqual() returns true iff two slices have the same strings
// in the same order. An empty slice is equal to a nil slice.
func StrListEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, ax := range a {
		if ax != b[i] {
			return false
		}
	}
	return true
}

// StrSetEqual() returns true iff two slices have the same set
// of strings, possibly in different order.
func StrSetEqual(a, b []string) bool {
	return StrListContainsAll(a, b) && StrListContainsAll(b, a)
}
