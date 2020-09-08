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
	ClientRequestModifiers      []Middleware                                                 `yaml:"-"`
	UpstreamModifiers           []func(r *http.Request) error                                `yaml:"-"`
	DownstreamModifiers         []func(w http.ResponseWriter, response *http.Response) error `yaml:"-"`
	hostMatch                   *regexp.Regexp                                               `yaml:"-"`
	pathMatch                   *regexp.Regexp                                               `yaml:"-"`
	cacheTimeOutDuration        time.Duration                                                `yaml:"-"`
	upstreamTimeoutDuration     time.Duration                                                `yaml:"-"`
	cacheMaxBodySizeInBytes     int64                                                        `yaml:"-"`
	upstreamURL                 *url.URL                                                     `yaml:"-"`
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
