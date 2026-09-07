package system

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// CommParams are the communication parameters for the http requests
type CommParams struct {
	Server      string
	Protocol    string // http, https
	URL         string
	Method      string // POST, PUT, GET
	Header      map[string]string
	IgnoreProxy string
	Timeout     int
	Retries     int
	Canonical   bool
}

// APIVersionsResponse contains the IMDS versiob response
type APIVersionsResponse struct {
	APIVersions []string `json:"apiVersions"`
}

// CommParamInit initialise the CommParams structure
func CommParamInit() CommParams {
	DebugLog("CommParamInit")
	commParams := CommParams{
		Server:      "",
		Protocol:    "http",
		URL:         "",
		Method:      "GET",
		Header:      make(map[string]string),
		IgnoreProxy: "no",
		Timeout:     2,
		Retries:     1,
		Canonical:   false,
	}
	DebugLog("CommParamInit -return commParams as '%+v'", commParams)
	return commParams
}

// RetrieveAPIVersion requests and extracts the API version from the server
// currently used for Azure
func RetrieveAPIVersion(params CommParams) string {
	DebugLog("RetrieveAPIVersion - params is '%+v'", params)
	var versRespons APIVersionsResponse
	apiVersion := ""
	vers := RetrievalRequest(params)
	//result: JSON string: "apiVersions":["2017-03-01","2017-04-02","2017-08-01","2017-10-01","2017-12-01", ... ,"2025-11-11"]}
	if len(vers) != 0 {
		err := json.Unmarshal([]byte(vers), &versRespons)
		DebugLog("vers is '%+v', versRespons is '%+v', err is '%v'", vers, versRespons, err)
		if err == nil && len(versRespons.APIVersions) > 0 {
			// extract latest version
			apiVersion = versRespons.APIVersions[len(versRespons.APIVersions)-1]
		}
	}
	DebugLog("RetrieveAPIVersion - return apiVersion as '%s'", apiVersion)
	return apiVersion
}

// responseMessage creates a logging message with the fields of the http
// response
func responseMessage(response *http.Response) string {
	DebugLog("responseMessage - response is '%+v'", response)
	msg := "response from server: nil"
	if response != nil {
		msg = fmt.Sprintf("response from server: Header '%+v', Status '%s', Body '%+v', ContentLength '%d', TransferEncoding '%+v', Close '%v', Uncompressed '%v', Trailer '%+v', Request '%+v', TLS '%+v'", response.Header, response.Status, response.Body, response.ContentLength, response.TransferEncoding, response.Close, response.Uncompressed, response.Trailer, response.Request, response.TLS)
	}
	DebugLog("responseMessage - return msg as '%s'", msg)
	return msg
}

// RetrievalRequest requests data from a HTTP server
func RetrievalRequest(params CommParams) string {
	DebugLog("RetrievalRequest - params is '%+v'", params)
	// create new request
	url := fmt.Sprintf("%s://%s/%s", params.Protocol, params.Server, params.URL)
	req, err := http.NewRequest(params.Method, url, nil)
	if err != nil {
		ErrorLog("failed to create %s request '%+v': %v", params.Protocol, req, err)
		return ""
	}

	// ANGI TODO - future: check and set/override header
	// add header, if needed
	if len(params.Header) != 0 {
		DebugLog("req.Header is '%+v'", req.Header)
		// for some requests we may need non-canonical keys,
		// so assigning the map directly.
		for key, val := range params.Header {
			if params.Canonical {
				req.Header.Add(key, val)
			} else {
				req.Header[key] = []string{val}
			}
		}
		DebugLog("req.Header is '%+v'", req.Header)
	}
	DebugLog("request to sent: '%+v'", req)

	// setup HTTP client
	client := &http.Client{
		Timeout: time.Duration(params.Timeout) * time.Second,
	}
	DebugLog("HTTP client setup is: '%+v'", client)

	// start the HTTP request
	var response *http.Response
	pass := false
	for retry := 0; retry < params.Retries; retry++ {
		response, err = client.Do(req)
		if err != nil {
			ErrorLog("failed to reach server %s: %v. Retry (%d) ...", params.Server, err, retry+1)
			ErrorLog(responseMessage(response))
			// Do NOT call response.Body.Close() if err != nil
			// (response is nil) -> will PANIC
			continue
		}
		// Ensure the request was successful
		if response.StatusCode != http.StatusOK {
			// ANGI TODO - detailed handle of status needed?
			// StatusCode >= 300 && StatusCode <= 399 ?? or others
			ErrorLog("server returned unexpected HTTP response - %d. Retry (%d) ...", response.StatusCode, retry+1)
			// Drain the body so the underlying TCP connection can
			// be reused. And close the body
			//io.Copy(io.Discard, response.Body)
			defer response.Body.Close()
			bdy, e := io.ReadAll(response.Body)
			ErrorLog(responseMessage(response))
			DebugLog("body is '%+v', as string '%s', e is '%v'", bdy, string(bdy), e)
			continue
		}
		InfoLog(responseMessage(response))
		pass = true
		break
	}
	if !pass {
		ErrorLog("Retrieval error, return.")
		return ""
	}
	defer response.Body.Close()

	// Read the raw response
	body, err := io.ReadAll(response.Body)
	if err != nil {
		ErrorLog("failed to read HTTP response body: '%s' - %v", strings.TrimSpace(string(body)), err)
		ErrorLog(responseMessage(response))
		return ""
	}
	DebugLog("RetrievalRequest - return body as '%s'", strings.TrimSpace(string(body)))
	return strings.TrimSpace(string(body))
}
