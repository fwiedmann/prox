package modifiers

import "net/http"

func SetProxyHTTPHeader(_ http.ResponseWriter, response *http.Response) error {
	response.Header.Set("x-hit-by-prox", "true")
	response.Header.Add("Connection", "x-hit-by-prox")
	return nil
}
