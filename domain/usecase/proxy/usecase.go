package proxy

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"sort"
	"strings"
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

type httpProxyUseCase struct {
	routerManager    route.Router
	cache            Cache
	port             uint16
	createHTTPClient func(route *route.Route) *http.Client
	accessLogEnabled bool
}

// NewUseCase creates a new proxy UseCase
func NewUseCase(manager route.Router, cache Cache, port uint16, accessLogEnabled bool, createHTTPClient func(route *route.Route) *http.Client) (UseCase, error) {
	if reflect.ValueOf(cache).Kind() == reflect.Ptr && reflect.ValueOf(cache).IsNil() {
		return nil, ErrInvalidCacheInterfaceValue
	}

	return &httpProxyUseCase{
		routerManager:    manager,
		cache:            cache,
		port:             port,
		accessLogEnabled: accessLogEnabled,
		createHTTPClient: createHTTPClient,
	}, nil
}

// ServeHTTP is the entrypoint for each incoming proxy request
func (u *httpProxyUseCase) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	if u.accessLogEnabled {
		log.WithFields(log.Fields{"ACCESS LOG": fmt.Sprintf("ACCESS %d", u.port)}).Infof("%+v", *r)
	}

	route, err := u.getRouteForRequest(r)
	if err != nil {
		http.Error(rw, ErrorStatusNotFound.Error(), http.StatusNotFound)
		return
	}
	chainMiddlewares(rootHandler{route: *route, cache: u.cache, Client: u.createHTTPClient(route)}.ServeHTTP, route.GetClientRequestModifiers()...).ServeHTTP(rw, r)
}

func (u *httpProxyUseCase) getRouteForRequest(r *http.Request) (*route.Route, error) {
	routes, err := u.routerManager.ListRoutes(r.Context())
	if err != nil {
		return nil, err
	}

	routeMatches := make([]*route.Route, 0)

	for _, route := range routes {
		if u.isRouteValidForRequest(route, strings.Split(r.Host, ":")[0], r.RequestURI) {
			routeMatches = append(routeMatches, route)
		}
	}

	if len(routeMatches) == 0 {
		return nil, ErrorNoMatchingRoute
	}

	sort.SliceStable(routeMatches, func(i, j int) bool {
		return routeMatches[i].Priority < routeMatches[j].Priority
	})

	return routeMatches[0], nil
}

func (u *httpProxyUseCase) isRouteValidForRequest(r *route.Route, host, path string) bool {
	if u.port != r.Port {
		return false
	}
	return r.IsHostnameMatching(host) && r.IsPathMatching(path)
}

type rootHandler struct {
	route route.Route
	cache Cache
	*http.Client
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
		resp, respErr = rh.Do(requestCopy)
		if respErr != nil {
			http.Error(rw, respErr.Error(), http.StatusInternalServerError)
			log.Errorf("upstream request error for route \"%s\" error: %v", rh.route.NameID, resp)
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

	rw.WriteHeader(resp.StatusCode)
	io.Copy(rw, resp.Body)
	close(stopChan)
}

func applyUpstreamModifiers(r *http.Request, route route.Route) error {
	for _, modFunc := range route.GetUpstreamModifiers() {
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
	for _, modFunc := range route.GetDownstreamModifiers() {
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
	request.Host = route.GetUpstreamURL().Host
	request.URL.Host = route.GetUpstreamURL().Host
	request.URL.Scheme = route.GetUpstreamURL().Scheme
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

// CreateHTTPClientForRoute configure a *http.Client based on a routes configure
func CreateHTTPClientForRoute(r *route.Route) *http.Client {
	return &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: r.UpstreamTLSValidation,
			},
		},
		Timeout: r.GetUpstreamTimeout(),
	}
}
