package proxy

import "net/http"

// UseCase defines the API for proxy http traffic
type UseCase interface {
	ServeHTTP(writer http.ResponseWriter, request *http.Request)
}
