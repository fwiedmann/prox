package route

import (
	"fmt"
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

// Middlewares
type Middlewares struct {
	HTTPSRedirect     bool `yaml:"https-redirect-enabled"`
	HTTPSRedirectPort int  `yaml:"https-redirect-port"`
	ForwardHostHeader bool `yaml:"forward-host-header"`
}

// Route entity contains all information of an proxy Router which can be used to configure proxy requests.
type Route struct {
	NameID                      NameID                                                       `yaml:"name"`
	CacheEnabled                bool                                                         `yaml:"cache-enabled"`
	CacheTimeOutDuration        string                                                       `yaml:"cache-timeout"`
	CacheMaxBodySizeInMegaBytes int64                                                        `yaml:"cache-max-body-size-in-mb"`
	CacheAllowedContentTypes    []string                                                     `yaml:"cache-allowed-content-types"`
	UpstreamURL                 string                                                       `yaml:"upstream-url"`
	UpstreamTimeoutDuration     string                                                       `yaml:"upstream-timeout"`
	UpstreamTLSValidation       bool                                                         `yaml:"upstream-skip-tls"`
	Priority                    uint                                                         `yaml:"priority"`
	Port                        uint16                                                       `yaml:"port"`
	Hostname                    RequestIdentifier                                            `yaml:"hostname"`
	HostnameRegexp              RequestIdentifier                                            `yaml:"hostname-regx"`
	Path                        RequestIdentifier                                            `yaml:"path"`
	PathRegexp                  RequestIdentifier                                            `yaml:"path-regx"`
	Middlewares                 Middlewares                                                  `yaml:"middlewares"`
	clientRequestModifiers      []Middleware                                                 `yaml:"-"`
	upstreamModifiers           []func(r *http.Request) error                                `yaml:"-"`
	downstreamModifiers         []func(w http.ResponseWriter, response *http.Response) error `yaml:"-"`
	hostMatch                   *regexp.Regexp                                               `yaml:"-"`
	pathMatch                   *regexp.Regexp                                               `yaml:"-"`
	cacheTimeOutDuration        time.Duration                                                `yaml:"-"`
	upstreamTimeoutDuration     time.Duration                                                `yaml:"-"`
	cacheMaxBodySizeInBytes     int64                                                        `yaml:"-"`
	upstreamURL                 *url.URL                                                     `yaml:"-"`
	httpClient                  *http.Client                                                 `yaml:"-"`
}

func (r *Route) GetHTTPClient() *http.Client {
	return r.httpClient
}

// String implements Stringer interface
func (r *Route) String() string {
	return fmt.Sprintf("%v", *r)
}

// GetClientRequestModifiers for the proxy request
func (r *Route) GetClientRequestModifiers() []Middleware {
	return r.clientRequestModifiers
}

// GetUpstreamModifiers for the proxy request
func (r *Route) GetUpstreamModifiers() []func(r *http.Request) error {
	return r.upstreamModifiers
}

// GetDownstreamModifiers for the proxy request
func (r *Route) GetDownstreamModifiers() []func(w http.ResponseWriter, response *http.Response) error {
	return r.downstreamModifiers
}

// GetUpstreamURL for the proxy request
func (r *Route) GetUpstreamURL() *url.URL {
	return r.upstreamURL
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
