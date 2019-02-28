package handler

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

func init() {
	prometheus.DefaultRegisterer.MustRegister(inFlightGauge, counter, responseSize, durationHistogram)
}

var (
	// Durations are partitioned by the HTTP method and use custom
	// buckets based on the expected request duration. ConstLabels are used
	// to set a handler label to mark which endpoint is being tracked.
	durationHistogram = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "A histogram of latencies for requests.",
			Buckets: []float64{.25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method"},
	)

	inFlightGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "http_in_flight_requests",
		Help: "A gauge of requests currently being served by the wrapped handler.",
	})

	counter = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "A counter for requests to the wrapped handler.",
		},
		[]string{"code", "method"},
	)

	responseSize = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "A histogram of response sizes for requests.",
			Buckets: []float64{1000, 3000, 7000, 10000, 20000, 50000},
		},
		[]string{},
	)
)

// Instrument instruments the provided http.Handler with prometheus metrics
func Instrument(h http.HandlerFunc) http.HandlerFunc {
	wrap := chain(httpLogMW, counterMW, durationMW, inFlightMW, responseSizeMW)
	return wrap(h)
}

type middleware func(handler http.HandlerFunc) http.HandlerFunc

func chain(mw ...middleware) middleware {
	return func(final http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			last := final
			for i := len(mw) - 1; i >= 0; i-- {
				last = mw[i](last)
			}
			last(w, r)
		}
	}
}
