package infra

import (
	"fmt"
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RouteStatusCode = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "prox_route_status_code",
		Help: "http status resp status code by prox route",
	}, []string{"status_code", "route"},
	)
	HTTPInMemCacheMaxSizeInBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prox_in_memeory_cache_max_size_in_bytes",
		Help: "max cache size in bytes",
	})

	HTTPInMemCacheCurrentSizeInBytes = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "prox_in_memeory_cache_curren_size_in_bytes",
		Help: "current cache size in bytes",
	})
)

//StartInfraHTTPEndpoint on a dedicated port which is not in use by the prox handlers
func StartInfraHTTPEndpoint(port int) error {
	mux := http.NewServeMux()
	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(prometheus.NewGoCollector(), prometheus.NewBuildInfoCollector(), RouteStatusCode, HTTPInMemCacheCurrentSizeInBytes, HTTPInMemCacheMaxSizeInBytes)
	mux.Handle("/metrics", promhttp.HandlerFor(metricsRegistry, promhttp.HandlerOpts{}))
	mux.HandleFunc("/health", HealthHandler)
	log.Debugf("Starting infra endpoint on port \"%d\". You can now open path /health or /metics", port)
	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

// HealthHandler handle http request on the /health endpoint
func HealthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, http.StatusText(http.StatusOK))
}
