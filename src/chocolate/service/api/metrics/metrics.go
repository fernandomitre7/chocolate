package metrics

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"chocolate/service/shared/logger"
	"chocolate/service/shared/reqcontext"
)

// Log is a Middleware function to log the Request Information after it completed
func Log(next http.Handler, name string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Forward Request to continue flow, and after it log everything
		next.ServeHTTP(w, r)
		startTime := reqcontext.GetStartTime(r)
		elapsed := time.Since(startTime)
		reqID := reqcontext.GetReqID(r)
		// Get Request IP Address
		requestClientIP := getRequestClientIP(r.RemoteAddr, r)
		// We will print the following information in the log file:
		// Col 1: Remote client address
		// Col 2: HTTP Method: GET, POST, PUT or DELETE
		// Col 3: Requested URI
		// Col 4: Request content length if any (in bytes)
		// Col 5: Status Code of the Response
		// Col 6: Length of response (in bytes)
		// Col 7: Execution time of the HTTP request

		/* status_code := context.GetStatusCode(r)
		payload_length := context.GetPayloadLen(r)
		response_length := context.GetResponseLength(r)
		req_id := context.GetReqId(r)*/
		logger.Infof("METRICS %s\t%s\t%s\t%s\t%v\t%v", reqID, requestClientIP, r.Method, r.URL.String(), startTime, elapsed)
	})
}

//  getRequestClientIP Retrieves the origin IP of an HTTP request using Headers 'X-Forwarded-For' & 'Forwarded'
func getRequestClientIP(connIP string, r *http.Request) string {
	var clientIP string
	var forwardedHeader string

	// By default, we take the origin's IP address of the current connection.
	clientIP = connIP

	// X-Forwarded-For case:
	// Syntax: X-Forwarded-For: <client>, <proxy1>, <proxy2>
	forwardedHeader = r.Header.Get("X-Forwarded-For")

	if len(forwardedHeader) > 0 {
		items := strings.Split(forwardedHeader, ",")
		// items[0] is always available and should contain the required value ...
		clientIP = items[0]
	} else {
		// Forwarded case:
		// Syntax: Forwarded: by=<identifier>; for=<identifier>; host=<host>; proto=<http|https>
		forwardedHeader = r.Header.Get("Forwarded")
		if len(forwardedHeader) > 0 {

			fwd, err := splitForwardHeader(forwardedHeader)

			if err != nil {
				logger.Warnf("metrics:getRequestClientIP: Error: %s", err)
			} else {
				originIP, ok := fwd["for"]
				if !ok {
					logger.Warn("metrics:getRequestClientIP: 'Forward' header doesn't contain 'host'.")
				} else {
					clientIP = originIP
				}
			}
		}
	}
	return clientIP
}

// splitForwardHeadersplits a string like: "for=192.0.2.60; proto=http; by=203.0.113.43"
// into a map like: { for => 192.0.2.60, proto => http, by => 203.0.113.43 }
func splitForwardHeader(header string) (fwdMap map[string]string, err error) {

	items := strings.Split(header, " ")
	fwdMap = make(map[string]string)

	for _, item := range items {
		keyValues := strings.SplitN(item, "=", 2)

		if len(keyValues) != 2 {
			err := fmt.Errorf("Malformed key/value pair on Forwarded Header: %s", header)
			return nil, err
		}

		valueSemiColon := strings.Split(keyValues[1], ";")
		fwdMap[keyValues[0]] = valueSemiColon[0]
	}

	return fwdMap, nil
}
