package altomsgs

import (
	"github.com/wdroome/go/wdrlib"
	"net/http"
	"net/url"
	"io"
	"bytes"
	"errors"
	"time"
	"strconv"
	"strings"
	"crypto/tls"
	_ "fmt"
	)

// HTTP header names.
const (
	CONTENT_TYPE_HDR = "Content-Type"
	CONTENT_LENGTH_HDR = "Content-Length"
	ACCEPT_HDR = "Accept"
	)

// AltoConn represents a connection to an ALTO server.
// To use, call LoadRootDir(uri) with the uri of the server's IRD.
// This reads that IRD, and any secondary IRDs, and creates 
// a ResourceSet with all the server's resources.
// After that, the methods NetworkMap(), CostMap(), etc,
// send the appropriate requests to the server and return the response.
type AltoConn struct {
	// HaveResources is true iff ResourceSet is valid.
	HaveResources bool
	
	// ResourceSet has the server's resources, and is created
	// by LoadRootDir().
	ResourceSet *ResourceSet
	
	// NetworkMapId is the ID of the network map which
	// will be used for NetworkMap and CostMap requests.
	// LoadRootDir sets it to the default network map.
	NetworkMapId string
	
	// ErrHandler() is called whenever an error occurs.
	// The method may log the error.
	ErrHandler func(errs []error)
	
	// Proxy is the url for the proxy, or nil.
	Proxy *url.URL

	// client defines the connection to the ALTO server.
	client *http.Client
	
	// transport is the Transport used by client.
	transport *http.Transport
}

// ServerResp describes the ALTO server's response to a request.
type ServerResp struct {
	// OkResp is the server's response to a successful request.
	// If the server returned an ALTO error response,
	// ErrorResp has the error, and OkResp is nil.
	OkResp AltoMsg
	
	// ErrorResp is the server's error response, or nil.
	ErrorResp *ErrorResp
	
	// Errors has any other errors which occured:
	// connection error, the server did not return
	// valid JSON, etc. If the server returns an ALTO
	// error, ErrorResp has the error and Errors is nil or 0-length.
	Errors []error
	
	// URI is the URI to which the request was sent.
	URI string
	
	// Request is the request body sent for a POST request.
	// nil for a GET request.
	Request AltoMsg
	
	// HaveResponse is true iff the server sent a response.
	// If HaveResponse is false, ContentType, ContentLength, etc,
	// have their default initial values, and should be ignored.
	HaveResponse bool
	
	// ContentType is the content type returned by the server.
	ContentType string
	
	// ContentLength has the length of the message returned by the server.
	ContentLength int64
	
	// Status has the HTTP status meesage returned by the server.
	// E.g., "200 OK".
	Status string
	
	// StatusCode has the integer HTTP status code (extracted from Status).
	StatusCode int
	
	// RespTime has the server's response time.
	RespTime time.Duration
}

// NewAltoConn() creates a new connection.
func NewAltoConn() *AltoConn {
	conn := AltoConn{}
	conn.setClient()
	return &conn
}

// LoadRootDir() reads a root IRD, and all secondary IRDs,
// and saves the ALTO server's resources in ResourceSet,
// replacing whatever was there before.
// Subsequent commands will use that ALTO server.
func (this *AltoConn) LoadRootDir(uri string) (time.Duration, []error) {
	this.setClient()
	if this.ResourceSet == nil {
		this.ResourceSet = NewResourceSet()
		this.ResourceSet.URI = uri
	}
	var totRespTime time.Duration = 0
	errs := this.addDirResources(uri, nil, nil, &totRespTime);
	this.NetworkMapId = this.ResourceSet.DefNetworkMapId
	this.HaveResources = len(this.ResourceSet.Resources) > 0
	return totRespTime, errs
}

// addDirResources() fetches an IRD, adds its resources to ResourceSet,
// and recursively calls itself on all secondary IRDs.
// pSecDirIds is a list of resource ids of all secondary IRDs
// which have been read. The function ignores any secondary IRDs
// whos IDs are on the list, and adds the IDs of any new IRDs.
// The function adds the server's response time to *pTotRespTime,
// if not nil.
func (this *AltoConn) addDirResources(uri string,
									  pSecDirIds *[]string,
									  prevErrs []error,
									  pTotRespTime *time.Duration) []error {
	if pSecDirIds == nil {
		pSecDirIds = &[]string{}
	}
	URI, err := url.Parse(uri)
	if err != nil {
		prevErrs = this.callErrHandler(prevErrs,
								"Invalid URI for IRD",
								http.MethodGet, uri, []error{err})
	}
	dir, serverResp := this.GetIRD(uri)
	if serverResp != nil {
		prevErrs = wdrlib.AppendErrors(prevErrs, serverResp.Errors)
		if pTotRespTime != nil {
			*pTotRespTime += serverResp.RespTime
		}
	}
	if dir != nil {
		errs := this.ResourceSet.AddResources(dir, URI)
		if len(errs) > 0 {
			prevErrs = this.callErrHandler(prevErrs,
									"Error in IRD resource",
									http.MethodGet, uri, []error{err})
		}
		for dirId, dirRes := range dir.Resources {
			if dirRes.MediaType == MT_DIRECTORY &&
						!wdrlib.StrListContains(*pSecDirIds, dirId) {
				res, ok := this.ResourceSet.Resources[dirId]
				if ok {
					*pSecDirIds = append(*pSecDirIds, dirId)
					prevErrs = this.addDirResources(res.URI.String(),
												pSecDirIds, prevErrs, pTotRespTime)
				}
			}
		}
	}
	return prevErrs
}

// GetIRD() reads and returns an IRD.
func (this *AltoConn) GetIRD(uri string) (*Directory, *ServerResp) {
	this.setClient()
	serverResp := this.SendReq(uri, []string{MT_DIRECTORY}, nil)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *Directory:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_DIRECTORY, http.MethodGet, uri)
		return nil, serverResp
	}
}

// NetworkMap() reads and returns the full Network Map with id NetworkMapId.
func (this *AltoConn) NetworkMap() (*NetworkMap, *ServerResp) {
	this.setClient()
	res, ok := this.ResourceSet.Resources[this.NetworkMapId]
	if !ok {
		errs := this.callErrHandler(nil,
								"No resource with id \"" + this.NetworkMapId + "\"",
								http.MethodGet, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	serverResp := this.SendReq(uri, []string{MT_NETWORK_MAP}, nil)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *NetworkMap:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_NETWORK_MAP, http.MethodGet, uri)
		return nil, serverResp
	}
}

// FilteredNetworkMap() returns a filtered Network Map
// for the indicated PIDs and address types.
func (this *AltoConn) FilteredNetworkMap(addrTypes []string,
										 pids []string) (*NetworkMap, *ServerResp) {
	this.setClient()
	res := this.ResourceSet.FindFilteredNetworkMap(this.NetworkMapId)
	if res == nil {
		errs := this.callErrHandler(nil,
								"No resource with id \"" + this.NetworkMapId + "\"",
								http.MethodPost, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	req := &NetworkMapFilter{AddrTypes: addrTypes, Pids: pids}
	serverResp := this.SendReq(uri, []string{MT_NETWORK_MAP}, req)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *NetworkMap:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_NETWORK_MAP, http.MethodPost, uri)
		return nil, serverResp
	}
}

// CostMap() reads and returns the full Cost Map for costType and network map NetworkMapId.
func (this *AltoConn) CostMap(costType CostType) (*CostMap, *ServerResp) {
	this.setClient()
	res := this.ResourceSet.FindCostMap(this.NetworkMapId, costType)
	if res == nil {
		errs := this.callErrHandler(nil,
								"No CostMap for " + costType.String() + " and netmap \"" +
											this.NetworkMapId + "\"",
								http.MethodGet, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	serverResp := this.SendReq(uri, []string{MT_COST_MAP}, nil)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *CostMap:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_COST_MAP, http.MethodGet, uri)
		return nil, serverResp
	}
}

// FilteredCostMap() returns a filtered CostMap
// for the indicated cost type, source and destination pids, and constraints.
func (this *AltoConn) FilteredCostMap(costType CostType,
									  srcs, dsts, constraints []string) (*CostMap, *ServerResp) {
	this.setClient()
	res := this.ResourceSet.FindFilteredCostMap(this.NetworkMapId,
										costType, len(constraints) > 0)
	if res == nil {
		errs := this.callErrHandler(nil,
								"No CostMap for " + costType.String() + " and netmap \"" +
											this.NetworkMapId + "\"",
								http.MethodPost, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	req := &CostMapFilter{Srcs: srcs, Dsts: dsts,
						 CostType: costType, Constraints: constraints}
	serverResp := this.SendReq(uri, []string{MT_COST_MAP}, req)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *CostMap:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_COST_MAP, http.MethodPost, uri)
		return nil, serverResp
	}
}

// EndpointCost() returns an EndpointCost
// for the indicated cost type, source and destination addresses, and constraints.
func (this *AltoConn) EndpointCost(costType CostType,
							srcs, dsts, constraints []string) (*EndpointCost, *ServerResp) {
	this.setClient()
	res := this.ResourceSet.FindEndpointCost(costType, len(constraints) > 0)
	if res == nil {
		errs := this.callErrHandler(nil,
								"No EndpointCost for " + costType.String() + "\"",
								http.MethodPost, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	req := &EndpointCostParams{Srcs: srcs, Dsts: dsts,
							   CostType: costType, Constraints: constraints}
	serverResp := this.SendReq(uri, []string{MT_ENDPOINT_COST}, req)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *EndpointCost:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_ENDPOINT_COST, http.MethodPost, uri)
		return nil, serverResp
	}
}

// EndpointProp() returns an EndpointProp
// for the indicated addresses and properties.
func (this *AltoConn) EndpointProp(addrs, propTypes []string) (*EndpointProp, *ServerResp) {
	this.setClient()
	res := this.ResourceSet.FindEndpointProp(propTypes)
	if res == nil {
		errs := this.callErrHandler(nil,
								"No EndpointProp for " + strings.Join(propTypes, " "),
								http.MethodPost, "", nil)
		return nil, &ServerResp{Errors: errs}
	}
	uri := res.URI.String()
	req := &EndpointPropParams{Endpoints: addrs, Properties: propTypes}
	serverResp := this.SendReq(uri, []string{MT_ENDPOINT_PROP}, req)
	if serverResp.OkResp == nil {
		return nil, serverResp
	}
	switch vv := serverResp.OkResp.(type) {
	case *EndpointProp:
		return vv, serverResp
	default:
		this.wrongRespType(serverResp, MT_ENDPOINT_PROP, http.MethodPost, uri)
		return nil, serverResp
	}
}

// SendReq() sends a request to "uri" and returns the response.
// "accept" has the media types the client expects;
// the function adds MT_ERROR if not in the list.
// If "req" is nil, use GET. If not, use POST
// and send "req" as the request message.
func (this *AltoConn) SendReq(uri string,
							  accept []string,
							  req AltoMsg) *ServerResp {
	this.setClient()
	serverResp := ServerResp{Errors: []error{},
							 URI: uri,
							 Request: req}
	var method string
	var sendContentType string = ""
	var sendData io.Reader = nil
	if req == nil {
		method = http.MethodGet
	} else {
		method = http.MethodPost
		sendContentType = req.MediaType()
		json, err := ToJsonBytes(req)
		if err != nil {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Error creating JSON for " + sendContentType,
							method, uri, []error{err})
			return &serverResp
		}
		sendData = bytes.NewBuffer(json)
	}
	httpReq, err := http.NewRequest(method, uri, sendData)
	if err != nil {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Error in http.NewRequest", method, uri, []error{err})
		return &serverResp
	}
	for _, mt := range accept {
		httpReq.Header.Add(ACCEPT_HDR, mt)
	}
	if !wdrlib.StrListContains(accept, MT_ERROR) {
		httpReq.Header.Add(ACCEPT_HDR, MT_ERROR)
	}
	if sendContentType != "" {
		httpReq.Header.Add(CONTENT_TYPE_HDR, sendContentType)
	}
	startTime := time.Now()
	httpResp, err := this.client.Do(httpReq)
	if err != nil {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Error in http.client.Do()", method, uri, []error{err})
		return &serverResp
	}
	serverResp.RespTime = time.Since(startTime)
	defer httpResp.Body.Close()
	
	serverResp.HaveResponse = true
	serverResp.Status = httpResp.Status
	serverResp.StatusCode = httpResp.StatusCode
	serverResp.ContentType = httpResp.Header.Get(CONTENT_TYPE_HDR)
	serverResp.ContentLength, _ = strconv.ParseInt(httpResp.Header.Get(CONTENT_LENGTH_HDR), 10, 64)
	if !(httpResp.StatusCode >= 200 && httpResp.StatusCode <= 299) {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Server returned HTTP status code " + strconv.Itoa(serverResp.StatusCode),
							method, uri, nil)
		if serverResp.ContentType != MT_ERROR {
			return &serverResp
		}
	}
	if serverResp.ContentType == "" {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"No " + CONTENT_TYPE_HDR + " in response",
							method, uri, nil)
		return &serverResp
	}
	resp, errs := NewAltoMsg(serverResp.ContentType, httpResp.Body, -1)
	if len(errs) > 0 {
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Cannot decode server response", method, uri, nil)
		return &serverResp
	}
	switch vv := resp.(type) {
	case *ErrorResp:
		serverResp.ErrorResp = vv
			serverResp.Errors = this.callErrHandler(
							serverResp.Errors,
							"Server returned error response, code=" + vv.Code,
							method, uri, nil)
	default:
		serverResp.OkResp = vv
	}
	return &serverResp
}

// SetTimeout() sets the timeout for a request. 0 means no timeout.
func (this *AltoConn) SetTimeout(timeout time.Duration) {
	this.setClient()
	this.client.Timeout = timeout
}

// Timeout() returns the timeout for a request. 0 means no timeout.
func (this *AltoConn) Timeout() time.Duration {
	this.setClient()
	return this.client.Timeout
}

// SetProxy() sets the proxy. Use "" to remove the proxy.
// The function returns nil if the proxy has been updated,
// or nil if uri is not a well-formed URL.
func (this *AltoConn) SetProxy(uri string) error {
	if uri == "" {
		this.Proxy = nil
		return nil
	} else {
		proxy, err := url.Parse(uri)
		if err == nil {
			this.Proxy = proxy
		}
		return err
	}
}

// SkipVerify() returns true iff we do not verify server certificates
// for TLS connections.
func (this *AltoConn) SkipVerify() bool {
	this.setClient()
	return this.transport.TLSClientConfig.InsecureSkipVerify
}

// SetSkipVerify() sets whether we verify server certificates
// for TLS connections. If skipVerify is true, do not verify.
// This should only be used for testing.
func (this *AltoConn) SetSkipVerify(skipVerify bool) {
	this.setClient()
	this.transport.TLSClientConfig.InsecureSkipVerify = skipVerify
}

// setClient() ensures that client and ResourceSet are not nil.
// If client is nil, the function sets it to the default HTTP client
// with a custom transport. If ResourceSet is nil, the function
// sets it to an empty set.
func (this *AltoConn) setClient() {
	if this.client == nil {
		this.transport = &http.Transport{
					Proxy: func (req *http.Request) (*url.URL, error) {
								return this.Proxy, nil
							},
					TLSClientConfig: &tls.Config{},
					}
		this.client = &http.Client{Transport: this.transport}
	}
	if this.ResourceSet == nil {
		this.ResourceSet = NewResourceSet()
	}
}

// callErrHandler() calls the custome error handler function
// on each error is "errors". The method then appends those errors
// to prevErrors, and returns the (possibly reallocated) slice.
func (this *AltoConn) callErrHandler(prevErrs []error,
									 descr, method, uri string,
									 errs []error) []error {
	if len(errs) == 0 {
		errs = []error{nil}
	}
	for _, err := range errs {
		msg := ""
		if err != nil {
			msg = " err=\"" + err.Error() + "\""
		}
		xerr := errors.New(descr + ": method=" + method +
							" uri=\"" + uri + "\"" + msg)
		if this.ErrHandler != nil {
			this.ErrHandler([]error{xerr})
		}
		prevErrs = append(prevErrs, xerr)
	}
	return prevErrs
}

// wrongRespType() is called when the ALTO server returns
// a message type other than expected one.
func (this *AltoConn) wrongRespType(serverResp *ServerResp,
									expected, method, uri string) {
	serverResp.Errors = this.callErrHandler(
								serverResp.Errors,
								"Incorrect response type: expected=" + expected +
								" actual=" + serverResp.OkResp.MediaType(), method, uri, nil)
	serverResp.OkResp = nil
}
