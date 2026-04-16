package obs

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type HTTPMetrics struct {
	Registry        *prometheus.Registry
	requestCount    *prometheus.CounterVec
	requestDuration *prometheus.HistogramVec
}

func NewHTTPMetrics(serviceName string) *HTTPMetrics {
	registry := prometheus.NewRegistry()
	buildInfo := promauto.With(registry).NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "service_build_info",
			Help: "Static build information.",
		},
		[]string{"service"},
	)
	buildInfo.WithLabelValues(serviceName).Set(1)

	requestCount := promauto.With(registry).NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests.",
		},
		[]string{"service", "method", "path", "status"},
	)

	requestDuration := promauto.With(registry).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"service", "method", "path", "status"},
	)

	return &HTTPMetrics{
		Registry:        registry,
		requestCount:    requestCount,
		requestDuration: requestDuration,
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.status = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func (m *HTTPMetrics) Middleware(serviceName string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			rec := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
			start := time.Now()

			next.ServeHTTP(rec, r)

			status := strconv.Itoa(rec.status)
			labels := []string{serviceName, r.Method, r.URL.Path, status}
			m.requestCount.WithLabelValues(labels...).Inc()
			m.requestDuration.WithLabelValues(labels...).Observe(time.Since(start).Seconds())
		})
	}
}
