package route

import (
	"net/http"
	"net/url"
	"regexp"
	"time"
)

// Middleware will be used to chain Middlewares before calling a root http.Handler.
type Middleware func(http.HandlerFunc) http.HandlerFunc

// NameID is an unique name for the Route.
type NameID string

// RequestIdentifier defines the validation query parameters which can be used to decide if incoming requests match with the Route.
type RequestIdentifier string

// Route entity contains all information of an router Router which can be used to configure proxy requests.
type Route struct {
	NameID                      NameID
	CacheEnabled                bool
	CacheTimeOutDuration        string
	CacheMaxBodySizeInMegaBytes int64
	CacheAllowedContentTypes    []string
	UpstreamURL                 *url.URL
	UpstreamTimeoutDuration     string
	UpstreamTLSValidation       bool
	Priority                    uint
	Port                        uint16
	Hostname                    RequestIdentifier
	HostnameRegexp              RequestIdentifier
	Path                        RequestIdentifier
	PathRegexp                  RequestIdentifier
	ClientRequestModifiers      []Middleware
	UpstreamModifiers           []func(r *http.Request) error
	DownstreamModifiers         []func(w http.ResponseWriter, response *http.Response) error
	hostMatch                   *regexp.Regexp
	pathMatch                   *regexp.Regexp
	cacheTimeOutDuration        time.Duration
	upstreamTimeoutDuration     time.Duration
	cacheMaxBodySizeInBytes     int64
}

// GetCacheMaxBodySizeInBytes return a validated bytes size
func (r *Route) GetCacheMaxBodySizeInBytes() int64 {
	return r.cacheMaxBodySizeInBytes
}

// GetCacheTimeOut returns a parsed duration
func (r *Route) GetCacheTimeOut() time.Duration {
	return r.cacheTimeOutDuration
}

// GetUpstreamTimeout returns a parsed duration
func (r *Route) GetUpstreamTimeout() time.Duration {
	return r.upstreamTimeoutDuration
}

// IsHostnameMatching check if h is valid hostname of the Route.
func (r *Route) IsHostnameMatching(h string) bool {
	return r.hostMatch.MatchString(h)
}

// IsPathMatching check if p is a valid path for of the Route.
func (r *Route) IsPathMatching(p string) bool {
	return r.pathMatch.MatchString(p)
}
