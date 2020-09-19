package cache

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/fwiedmann/prox/domain/entity/route"
)

func TestHTTPInMemoryCache_Get(t *testing.T) {
	t.Parallel()
	type fields struct {
		store               map[string]response
		mtx                 sync.RWMutex
		maxCacheSizeInBytes int64
		cacheSizeInBytes    int64
	}
	type args struct {
		route   route.Route
		request *http.Request
	}
	tests := []struct {
		name               string
		fields             fields
		args               args
		wantNil            bool
		cacheSizeAfterExec int64
	}{
		{
			name: "NilResponse",
			fields: fields{
				store:               map[string]response{},
				mtx:                 sync.RWMutex{},
				maxCacheSizeInBytes: 0,
				cacheSizeInBytes:    2,
			},
			args: args{
				route:   route.Route{NameID: "test-route"},
				request: &http.Request{Host: "test.com", RequestURI: "/hello", ContentLength: 1},
			},
			wantNil:            true,
			cacheSizeAfterExec: 2,
		},
		{
			name: "ValidReturn",
			fields: fields{
				store:               map[string]response{"test-route-test.com-/hello": {contentLength: 1, header: map[string][]string{httpInMemoryCacheHeader: {"true"}}}},
				mtx:                 sync.RWMutex{},
				maxCacheSizeInBytes: 0,
				cacheSizeInBytes:    2,
			},
			args: args{
				route:   route.Route{NameID: "test-route"},
				request: &http.Request{Host: "test.com", RequestURI: "/hello", ContentLength: 1},
			},
			wantNil:            false,
			cacheSizeAfterExec: 1,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			hc := &HTTPInMemoryCache{
				store:               tt.fields.store,
				mtx:                 tt.fields.mtx, //nolint
				maxCacheSizeInBytes: tt.fields.maxCacheSizeInBytes,
				cacheSizeInBytes:    tt.fields.cacheSizeInBytes,
			}
			got := hc.Get(tt.args.route, tt.args.request)
			if got != nil && tt.wantNil {
				t.Errorf("Get() = %v, wantNil %v", got, tt.wantNil)
			}

			if tt.cacheSizeAfterExec != hc.cacheSizeInBytes {
				t.Error("did not update cacheSizeAfterExec ")
			}

			if got != nil {
				if got.Header.Get(httpInMemoryCacheHeader) == "" {
					t.Error("Get() did not set the cache header")
				}

			}

		})
	}
}

func TestHTTPInMemoryCache_isValidateSave(t *testing.T) {
	t.Parallel()
	type fields struct {
		store               map[string]response
		mtx                 sync.RWMutex
		maxCacheSizeInBytes int64
		cacheSizeInBytes    int64
	}
	type args struct {
		route   route.Route
		request *http.Request
		resp    *http.Response
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name:   "invalidMethod",
			fields: fields{},
			args: args{
				request: &http.Request{Method: http.MethodPut},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 15, Hostname: "docker.com"},
			},
			want: false,
		},
		{
			name:   "invalidMethod",
			fields: fields{},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 400},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 15, Hostname: "docker.com"},
			},
			want: false,
		},
		{
			name: "cacheIsFull",
			fields: fields{
				maxCacheSizeInBytes: 10,
			},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 200, ContentLength: 100},
				route:   route.Route{NameID: "test-route", Hostname: "docker.com"},
			},
			want: false,
		},
		{
			name: "toBigBodySize",
			fields: fields{
				maxCacheSizeInBytes: 1110000000,
			},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 200, ContentLength: 110000000},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 10, Hostname: "docker.com"},
			},
			want: false,
		},
		{
			name: "invalidContentType",
			fields: fields{
				maxCacheSizeInBytes: 20,
			},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 200, ContentLength: 10, Header: map[string][]string{httpContentTypeHeader: {"yaml"}}},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 15, Hostname: "docker.com", CacheAllowedContentTypes: []string{"json"}},
			},
			want: false,
		},
	}
	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {
			hc := &HTTPInMemoryCache{
				store:               tt.fields.store,
				mtx:                 tt.fields.mtx, //nolint
				maxCacheSizeInBytes: tt.fields.maxCacheSizeInBytes,
				cacheSizeInBytes:    tt.fields.cacheSizeInBytes,
			}

			m := route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute)

			if err := m.CreateRoute(context.Background(), &tt.args.route); err != nil {
				t.Error(err)
				return
			}

			routes, err := m.ListRoutes(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			var routeToUse *route.Route
			for _, r := range routes {
				if r.NameID == tt.args.route.NameID {
					routeToUse = r
				}
			}
			if routeToUse == nil {
				t.Error("route is empty")
				return
			}

			if got := hc.isValidateSave(*routeToUse, tt.args.request, tt.args.resp); got != tt.want {
				t.Errorf("isValidateSave() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestHTTPInMemoryCache_Save(t *testing.T) {
	t.Parallel()
	type fields struct {
		store               map[string]response
		mtx                 sync.RWMutex
		maxCacheSizeInBytes int64
		cacheSizeInBytes    int64
	}
	type args struct {
		route   route.Route
		request *http.Request
		resp    *http.Response
	}
	tests := []struct {
		name                string
		fields              fields
		args                args
		storeCountAfterSave int
		cacheSizeAfterSave  int64
	}{

		{
			name: "InvalidSave",
			fields: fields{
				store:               map[string]response{},
				mtx:                 sync.RWMutex{},
				maxCacheSizeInBytes: 0,
				cacheSizeInBytes:    0,
			},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 200, ContentLength: 15000},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 10, Hostname: "docker.com"},
			},
			storeCountAfterSave: 0,
			cacheSizeAfterSave:  0,
		},
		{
			name: "ValidSave",
			fields: fields{
				store:               map[string]response{},
				mtx:                 sync.RWMutex{},
				maxCacheSizeInBytes: 1000000000000,
				cacheSizeInBytes:    0,
			},
			args: args{
				request: &http.Request{Method: http.MethodGet},
				resp:    &http.Response{StatusCode: 200, ContentLength: 15000, Body: ioutil.NopCloser(bytes.NewBuffer([]byte{1}))},
				route:   route.Route{NameID: "test-route", CacheMaxBodySizeInMegaBytes: 1000, Hostname: "docker.com"},
			},
			storeCountAfterSave: 1,
			cacheSizeAfterSave:  15000,
		},
	}

	for _, tt := range tests { //nolint
		t.Run(tt.name, func(t *testing.T) {

			m := route.NewManager(route.NewInMemRepo(), route.CreateHTTPClientForRoute)

			if err := m.CreateRoute(context.Background(), &tt.args.route); err != nil {
				t.Error(err)
				return
			}

			routes, err := m.ListRoutes(context.Background())
			if err != nil {
				t.Error(err)
				return
			}
			var routeToUse *route.Route
			for _, r := range routes {
				if r.NameID == tt.args.route.NameID {
					routeToUse = r
				}
			}
			if routeToUse == nil {
				t.Error("route is empty")
				return
			}

			hc := &HTTPInMemoryCache{
				store:               tt.fields.store,
				mtx:                 tt.fields.mtx, //nolint
				maxCacheSizeInBytes: tt.fields.maxCacheSizeInBytes,
				cacheSizeInBytes:    tt.fields.cacheSizeInBytes,
			}
			hc.Save(*routeToUse, tt.args.request, tt.args.resp)

			if tt.cacheSizeAfterSave != hc.cacheSizeInBytes {
				t.Errorf("Save() cache size want %d, got %d", tt.cacheSizeAfterSave, hc.cacheSizeInBytes)
			}

			if tt.storeCountAfterSave != len(hc.store) {
				t.Errorf("Save() store size want %d, got %d", tt.storeCountAfterSave, len(hc.store))
			}

		})
	}
}
