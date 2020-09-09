package modifiers

import (
	"net"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// ForwardHost will set the X-Forwarded-For header with the clients remote address
func ForwardHost(r *http.Request) error {
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		log.Error(err)
	}
	r.Header.Set("X-Forwarded-For", host)
	return nil
}
