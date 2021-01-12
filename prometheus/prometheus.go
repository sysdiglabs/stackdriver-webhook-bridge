package prometheus

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	log "github.com/sirupsen/logrus"
)

var (
	httpServer *http.Server
)

func ExposeMetricsEndpoint(port int) {
	// Start Prometheus metrics endpoint
	router := chi.NewRouter()
	router.Method(http.MethodGet, "/metrics", promhttp.Handler())
	hostPort := fmt.Sprintf(":%d", port)
	log.Infof("Starting metrics server on %s...", hostPort)

	httpServer = &http.Server {
		Addr: hostPort,
		Handler: router,
	}

	if err := httpServer.ListenAndServe(); err != http.ErrServerClosed {
		log.Fatalf("Cannot start metrics server: %v", err)
	}
}
