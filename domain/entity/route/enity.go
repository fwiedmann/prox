package route

import (
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Middleware // Todo add description
type Middleware func(http.HandlerFunc) http.HandlerFunc

// NameID
type NameID string

type RequestIdentifier string

// Route entity // Todo add description
type Route struct {
	Name                   NameID
	UpstreamURL            *url.URL
	UpstreamTimeout        time.Duration
	UpstreamTLSValidation  bool
	Priority               uint
	Port                   uint16
	Hostname               RequestIdentifier
	HostnameRegexp         RequestIdentifier
	Path                   RequestIdentifier
	PathRegexp             RequestIdentifier
	ClientRequestModifiers []Middleware
	UpstreamModifiers      []func(r *http.Request) error
	DownstreamModifiers    []func(w http.ResponseWriter, response *http.Response) error
	hostMatch              *regexp.Regexp
	pathMatch              *regexp.Regexp
}

func (r *Route) IsHostnameMatching(h string) bool {
	return r.hostMatch.MatchString(h)
}

func (r *Route) IsPathMatching(p string) bool {
	return r.pathMatch.MatchString(p)
}
