package cache

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"sync"
	"time"

	"github.com/fwiedmann/prox/domain/entity/route"
)

type response struct {
	body          []byte
	contentLength int64
	header        http.Header
	statusCode    int
	status        string
}

const megaBytesToBytesMultiplier = 1e+6
const httpInMemoryCacheHeader = "x-cached-by-prox"
const httpContentTypeHeader = "Content-Type"

// NewHTTPInMemoryCache creates a new http cache which stores http responses in memory
func NewHTTPInMemoryCache(maxCacheSizeInMegaBytes int64) *HTTPInMemoryCache {
	return &HTTPInMemoryCache{
		store:               make(map[string]response),
		maxCacheSizeInBytes: maxCacheSizeInMegaBytes * megaBytesToBytesMultiplier,
	}
}

// HTTPInMemoryCache stores http responses in memory for a defined duration
type HTTPInMemoryCache struct {
	store               map[string]response
	mtx                 sync.RWMutex
	maxCacheSizeInBytes int64
	cacheSizeInBytes    int64
}

// Get return a stored in memory response. If no response was found nil will be returned
func (hc *HTTPInMemoryCache) Get(route route.Route, request *http.Request) *http.Response {
	hc.mtx.RLock()
	defer hc.mtx.RUnlock()

	if val, ok := hc.store[buildID(route, request)]; ok {
		if val.contentLength > 0 {
			hc.cacheSizeInBytes -= val.contentLength
		}
		val.header.Set(httpInMemoryCacheHeader, "true")

		return &http.Response{
			Body:          ioutil.NopCloser(bytes.NewBuffer(val.body)),
			ContentLength: val.contentLength,
			Header:        val.header,
			StatusCode:    val.statusCode,
			Status:        val.status,
		}
	}
	return nil
}

// Save a http request with its body in memory. The route.NameID, http.Request.Hostname and http.Request.RequestURI will be used to generate an ID.
func (hc *HTTPInMemoryCache) Save(route route.Route, request *http.Request, resp *http.Response) {
	if !hc.isValidateSave(route, request, resp) {
		return
	}

	hc.mtx.Lock()
	defer hc.mtx.Unlock()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))
		return
	}

	resp.Body = ioutil.NopCloser(bytes.NewBuffer(body))

	id := buildID(route, request)
	hc.store[id] = response{
		body:          body,
		contentLength: resp.ContentLength,
		header:        resp.Header,
		statusCode:    resp.StatusCode,
		status:        resp.Status,
	}

	if resp.ContentLength > 0 {
		hc.cacheSizeInBytes += resp.ContentLength
	}

	go hc.deleteStoredResponseAfterTimeout(id, route.GetCacheTimeOut())
}

func (hc *HTTPInMemoryCache) isValidateSave(route route.Route, request *http.Request, resp *http.Response) bool {
	if request.Method != http.MethodGet {
		return false
	}

	if !isValidResponseCode(resp.StatusCode) {
		return false
	}

	if hc.maxCacheSizeInBytes-hc.cacheSizeInBytes <= resp.ContentLength && hc.maxCacheSizeInBytes != -1 {
		return false
	}

	if route.GetCacheMaxBodySizeInBytes() < resp.ContentLength && route.GetCacheMaxBodySizeInBytes() != -1 {
		return false
	}

	if len(route.CacheAllowedContentTypes) != 0 && !isValidContentType(resp, route.CacheAllowedContentTypes) {
		return false
	}
	return true
}

func (hc *HTTPInMemoryCache) deleteStoredResponseAfterTimeout(id string, duration time.Duration) {
	time.Sleep(duration)
	hc.mtx.Lock()
	defer hc.mtx.Unlock()
	hc.cacheSizeInBytes -= hc.store[id].contentLength
	delete(hc.store, id)
}

func isValidResponseCode(code int) bool {
	switch code {
	case http.StatusOK:
		return true
	case http.StatusNotModified:
		return true
	default:
		return false
	}
}

func isValidContentType(r *http.Response, allowedContentTypes []string) bool {
	requestContentType := r.Header.Get(httpContentTypeHeader)
	if requestContentType == "" {
		return false
	}

	for _, contentType := range allowedContentTypes {
		if contentType == requestContentType {
			return true
		}
	}
	return false
}

func buildID(route route.Route, clientRequest *http.Request) string {
	path := "/"
	if clientRequest.RequestURI != "" {
		path = clientRequest.RequestURI
	}
	return fmt.Sprintf("%s-%s-%s", route.NameID, clientRequest.Host, path)
}
