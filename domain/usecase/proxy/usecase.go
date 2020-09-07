package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net"
	"net/http"
	"reflect"
	"sort"
	"time"

	log "github.com/sirupsen/logrus"

	"github.com/fwiedmann/prox/domain/entity/route"
)

var (
	ErrorNoMatchingRoute           = errors.New("no matching route found")
	ErrorStatusNotFound            = errors.New("404 - Not Found")
	ErrorStatusInternalServerError = errors.New("500 - Internal Server Error")
	ErrInvalidCacheInterfaceValue  = errors.New("cache is not allowed to be nil or a pointer")
)

// Cache defines a API for caching *http.Response
type Cache interface {
	Get(route route.Route, request *http.Request) *http.Response
	Save(route route.Route, request *http.Request, response *http.Response)
}

type useCase struct {
	routerManager route.Manager
	cache         Cache
	port          uint16
}

// NewUseCase creates a new proxy UseCase
func NewUseCase(manager route.Manager, cache Cache, port uint16) (UseCase, error) {
	if reflect.ValueOf(cache).Kind() == reflect.Ptr && reflect.ValueOf(cache).IsNil() {
		return nil, ErrInvalidCacheInterfaceValue
	}

	return &useCase{
		routerManager: manager,
		cache:         cache,
		port:          port,
	}, nil
}

// ServeHTTP is the entrypoint for each incoming proxy request
func (u *useCase) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	route, err := u.getRouteForRequest(r)
	if err != nil {
		http.Error(rw, ErrorStatusNotFound.Error(), http.StatusNotFound)
		return
	}
	chainMiddlewares(rootHandler{route: route, cache: u.cache}.ServeHTTP, route.ClientRequestModifiers...).ServeHTTP(rw, r)
}

func (u *useCase) getRouteForRequest(r *http.Request) (route.Route, error) {
	routes, err := u.routerManager.ListRoutes(r.Context())
	if err != nil {
		return route.Route{}, err
	}

	routeMatches := make([]route.Route, 0)

	for _, route := range routes {
		host, _, err := net.SplitHostPort(r.Host)
		if err != nil {
			log.Errorf("could not parse host for request %+v, error: %s", *r, err)
			continue
		}

		if u.isRouteValidForRequest(route, host, r.RequestURI) {
			routeMatches = append(routeMatches, *route)
		}
	}

	if len(routeMatches) == 0 {
		return route.Route{}, ErrorNoMatchingRoute
	}

	sort.SliceStable(routeMatches, func(i, j int) bool {
		return routeMatches[i].Priority < routeMatches[j].Priority
	})

	return routeMatches[0], nil
}

func (u *useCase) isRouteValidForRequest(r *route.Route, host, path string) bool {
	if u.port != r.Port {
		return false
	}
	return r.IsHostnameMatching(host) && r.IsPathMatching(path)
}

type rootHandler struct {
	route route.Route
	cache Cache
}

// ServeHTTP is the main proxy handler
func (rh rootHandler) ServeHTTP(rw http.ResponseWriter, r *http.Request) {

	var resp *http.Response
	if rh.route.CacheEnabled {
		resp = rh.cache.Get(rh.route, r)
	}

	stopChan := make(chan struct{})
	if resp == nil {
		requestCopy := r.Clone(r.Context())
		if err := applyUpstreamModifiers(requestCopy, rh.route); err != nil {
			http.Error(rw, ErrorStatusInternalServerError.Error(), http.StatusInternalServerError)
			log.Errorf("could not apply upstream request modifiers for route \"%s\" error: %s", rh.route.NameID, err)
			return
		}

		configureRequestForUpstream(requestCopy, rh.route)

		var respErr error
		resp, respErr = configureHTTPClientForRoute(rh.route).Do(requestCopy)
		if respErr != nil {
			http.Error(rw, respErr.Error(), http.StatusInternalServerError)
			log.Errorf("upstream request error for route \"%s\" error: %s", rh.route.NameID, resp)
			return
		}
		defer resp.Body.Close()

		if err := applyDownstreamModifiers(r.Context(), rw, resp, rh.route); err != nil {
			http.Error(rw, ErrorStatusInternalServerError.Error(), http.StatusInternalServerError)
			log.Errorf("could not down upstream request modifiers for route \"%s\" error: %s", rh.route.NameID, err)
			return
		}

		if rh.route.CacheEnabled {
			rh.cache.Save(rh.route, r, resp)
		}

	}

	configureHeadersForClientFromResponseHeaders(rw.Header(), resp.Header)

	if isRespIsBuffered(resp.TransferEncoding) {
		go flushResponse(stopChan, rw)
	}
	io.Copy(rw, resp.Body)
	close(stopChan)
}

func applyUpstreamModifiers(r *http.Request, route route.Route) error {
	for _, modFunc := range route.UpstreamModifiers {
		if err := modFunc(r); err != nil {
			return err
		}
		if err := r.Context().Err(); err != nil {
			return err
		}
	}
	return nil
}

func applyDownstreamModifiers(ctx context.Context, w http.ResponseWriter, response *http.Response, route route.Route) error {
	for _, modFunc := range route.DownstreamModifiers {
		if err := modFunc(w, response); err != nil {
			return err
		}
		if err := ctx.Err(); err != nil {
			return err
		}
	}
	return nil
}

func configureRequestForUpstream(request *http.Request, route route.Route) {
	request.Host = route.UpstreamURL.Host
	request.URL.Host = route.UpstreamURL.Host
	request.URL.Scheme = route.UpstreamURL.Scheme
	request.RequestURI = ""
}

func configureHeadersForClientFromResponseHeaders(clientResponseHeader, upstreamResponseHeader http.Header) {
	for key, headerValues := range upstreamResponseHeader {
		for _, value := range headerValues {
			clientResponseHeader.Add(key, value)
		}
	}
	clientResponseHeader.Set("cache-control", "max-age=0, private, must-revalidate, no-store")
}

func chainMiddlewares(rootHandler http.HandlerFunc, middlewares ...route.Middleware) http.HandlerFunc {
	if len(middlewares) < 1 {
		return rootHandler
	}
	chain := rootHandler
	for i := len(middlewares) - 1; i >= 0; i-- {
		chain = middlewares[i](chain)
	}
	return chain
}

func configureHTTPClientForRoute(r route.Route) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: r.UpstreamTLSValidation,
			},
		},
		Timeout: r.GetUpstreamTimeout(),
	}
}

func flushResponse(c <-chan struct{}, rw http.ResponseWriter) {
	for {
		select {
		case <-time.Tick(5 * time.Millisecond):
			flusher, ok := rw.(http.Flusher)
			if !ok {
				return
			}
			flusher.Flush()
		case <-c:
			return
		}
	}
}

func isRespIsBuffered(transferEncoding []string) bool {
	for _, entry := range transferEncoding {
		if entry == "chunked" {
			return true
		}
	}
	return false
}
