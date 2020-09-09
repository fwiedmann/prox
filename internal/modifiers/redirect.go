package modifiers

import (
	"fmt"
	"net/http"
	"strings"
)

// HTTPSRedirect configuration
type HTTPSRedirect struct {
	port int
}

// NewHTTPSRedirect init a new HTTPSRedirect handler
func NewHTTPSRedirect(port int) HTTPSRedirect {
	return HTTPSRedirect{port: port}
}

// Redirect Will redirect the client to scheme https and the configured HTTPSRedirect.Port
func (hr HTTPSRedirect) Redirect(next http.HandlerFunc) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {

		if request.TLS != nil {
			next.ServeHTTP(writer, request)
			return
		}
		http.Redirect(writer, request, fmt.Sprintf("https://%s:%d", strings.Split(request.Host, ":")[0], hr.port), http.StatusFound)
	}
}
