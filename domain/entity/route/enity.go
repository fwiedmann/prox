package route

import (
	"net/http"
	"net/url"
	"regexp"
	"time"

	"github.com/fwiedmann/prox/domain/entity"
)

// Middleware // Todo add description
type Middleware func(http.HandlerFunc) http.HandlerFunc

// Route entity // Todo add description
type Route struct {
	ID                    entity.ID
	UpstreamURL           *url.URL
	UpstreamTimeout       time.Duration
	UpstreamTLSValidation bool
	// Todo make matcher private, add Host,Path string and HostRegexExpression,PathRegexExpression
	// Todo add methods which will call the hostmatch, pathmatch.Matchstring()
	HostMatch              *regexp.Regexp
	PathMatch              *regexp.Regexp
	Priority               int
	ClientRequestModifiers []Middleware
	UpstreamModifiers      []func(r *http.Request) error
	DownstreamModifiers    []func(w http.ResponseWriter, response *http.Response) error
	Port                   int
}
