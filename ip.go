package gottle

import (
	"net/http"
	"strings"
)

var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

//RealIP is an IPProvider implementation that fetches the ip of an HTTP
//request by inspecting the "X-Forwarded-For" or "X-Real-IP" headers
//This should only be used when you have a reverse proxy in place.
type RealIP struct{}

//IP returns the ip associated with the request
// Ported from pressly/chi
func (re *RealIP) IP(r *http.Request) string {

	var ip string

	if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ", ")

		if i == -1 {
			i = len(xff)
		}

		ip = xff[:i]

	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	}

	return ip
}

//NewRealIP returns an instance of RealIP
func NewRealIP() *RealIP {
	return &RealIP{}
}

//RemoteIP is an IPProvider that fetches the IP of the request directly
// from the `RemoteAddr` of the Request
//This is extremely unreliable.
// Go sets the `RemoteAddr` to "IP:port" before your app handlers are called
//So you must be setting it to the right value in a middleware or something
//Do look at the RealIP implementation or consider writing your own IPProvider
type RemoteIP struct{}

func (rip *RemoteIP) IP(r *http.Request) string {
	return r.RemoteAddr
}

//NewRemoteIP returns an instance of the RemoteIP implementation of IPProvider
func NewRemoteIP() *RemoteIP {
	return &RemoteIP{}
}
