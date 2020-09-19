package cache

import (
	"net/http"

	"github.com/fwiedmann/prox/domain/entity/route"
)

// Empty implementation for the proxy.Cache interface
type Empty struct{}

// Get will always return nil
func (Empty) Get(_ route.Route, _ *http.Request) *http.Response {
	return nil
}

// Save will always discard the request
func (Empty) Save(route route.Route, _ *http.Request, response *http.Response) {
}
