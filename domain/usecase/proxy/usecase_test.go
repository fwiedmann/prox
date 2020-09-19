package proxy

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/fwiedmann/prox/internal/cache"

	"github.com/fwiedmann/prox/domain/entity/route"
)

type tripper struct {
	resp *http.Response
	err  error
}

func (t tripper) RoundTrip(_ *http.Request) (*http.Response, error) {
	return t.resp, t.err
}

type clientCreator struct {
	respCode         int
	body             []byte
	header           map[string][]string
	respErr          error
	TransferEncoding string
}

func (c clientCreator) CreateFakeHTTPClient(_ *route.Route) *http.Client {
	return &http.Client{
		Transport: tripper{
			resp: &http.Response{StatusCode: c.respCode, Body: ioutil.NopCloser(bytes.NewBuffer(c.body)), Header: c.header, TransferEncoding: []string{c.TransferEncoding}},
			err:  c.respErr,
		},
	}
}

func Test_httpProxyUseCase_ServeHTTP(t *testing.T) {
	type initFields struct {
		routes []*route.Route
	}

	type fields struct {
		cache            Cache
		createHTTPClient func(route *route.Route) *http.Client
	}
	tests := []struct {
		name string
		initFields
		fields         fields
		requestPath    string
		wantStatusCode int
		wantBody       string
	}{
		{
			name: "ValidRequest",
			initFields: initFields{
				routes: []*route.Route{{
					NameID:       "test-route",
					CacheEnabled: true,
					UpstreamURL:  "test.localhost",
					Path:         "/hello",
				}},
			},
			fields: fields{
				cache:            cache.NewHTTPInMemoryCache(-1),
				createHTTPClient: clientCreator{body: []byte("ok"), header: map[string][]string{"test": {"test"}}, respCode: 200}.CreateFakeHTTPClient,
			},
			wantStatusCode: 200,
			wantBody:       "ok",
			requestPath:    "/hello",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			m := route.NewManager(route.NewInMemRepo(), tt.fields.createHTTPClient)

			for _, r := range tt.initFields.routes {
				if err := m.CreateRoute(context.Background(), r); err != nil {
					t.Error(err)
					return
				}
			}
			u := &httpProxyUseCase{
				routerManager: m,
				cache:         tt.fields.cache,
			}

			server := httptest.NewServer(u)
			defer server.Close()

			_, port, _ := net.SplitHostPort(server.URL)

			parsedPort, _ := strconv.Atoi(port)

			u.port = uint16(parsedPort)

			resp, err := http.Get(server.URL + tt.requestPath)
			if err != nil {
				t.Error(err)
				return
			}

			if resp.StatusCode != tt.wantStatusCode {
				t.Errorf("response code got %d, want %d", resp.StatusCode, tt.wantStatusCode)
			}

			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Errorf(err.Error())
				return
			}
			if !strings.Contains(string(body), tt.wantBody) {
				t.Errorf("response body is %s, want %s", string(body), tt.wantBody)
			}
		})
	}
}
