package altomsgs

/*
 * Error types. All implement error.
 */

// CIDRError means a CIDR is invalid.
type CIDRError struct {
	CIDR string
	Err string
}
var _ error = CIDRError{}

func (this CIDRError) Error() string {
	return "Invalid CIDR '" + this.CIDR + "': " + this.Err
}

// JSONTypeError means a JSON field has the wrong type.
type JSONTypeError struct {
	Path string
	Err string
}
var _ error = JSONTypeError{}

func (this JSONTypeError) Error() string {
	return "Wrong type '" + this.Path + "': " + this.Err
}
