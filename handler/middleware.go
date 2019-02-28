package handler

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/digitalocean/metis/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func inFlightMW(handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerInFlight(inFlightGauge, handler).ServeHTTP
}

func counterMW(handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerCounter(counter, handler)
}

func responseSizeMW(handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerResponseSize(responseSize, handler).ServeHTTP
}

func durationMW(handler http.HandlerFunc) http.HandlerFunc {
	return promhttp.InstrumentHandlerDuration(durationHistogram, handler)
}

// httpLogMW incoming requests, including response status.
func httpLogMW(h http.HandlerFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		o := &responseObserver{ResponseWriter: w}
		h.ServeHTTP(o, r)
		addr := r.RemoteAddr
		if i := strings.LastIndex(addr, ":"); i != -1 {
			addr = addr[:i]
		}

		if o.status == 0 {
			o.status = 200
		}

		var logmethod func(msg string, params ...interface{}) = log.Info
		if o.status%100 == 5 {
			logmethod = log.Error
		}
		logmethod("%s - - [%s] %q %d %d %q %q",
			addr,
			time.Now().Format("02/Jan/2006:15:04:05 -0700"),
			fmt.Sprintf("%s %s %s", r.Method, r.URL, r.Proto),
			o.status,
			o.written,
			r.Referer(),
			r.UserAgent())
	})
}

type responseObserver struct {
	http.ResponseWriter
	status      int
	written     int64
	wroteHeader bool
}

func (o *responseObserver) Write(p []byte) (n int, err error) {
	if !o.wroteHeader {
		o.WriteHeader(http.StatusOK)
	}
	n, err = o.ResponseWriter.Write(p)
	o.written += int64(n)
	return
}

func (o *responseObserver) WriteHeader(code int) {
	o.ResponseWriter.WriteHeader(code)
	if o.wroteHeader {
		return
	}
	o.wroteHeader = true
	o.status = code
}
