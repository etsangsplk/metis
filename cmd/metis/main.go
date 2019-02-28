package main

import (
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/digitalocean/metis/handler"
	"github.com/digitalocean/metis/log"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/tsdb"
)

func main() {
	if err := InitConfig(); err != nil {
		log.Fatal("init failed: %v", err)
	}

	db, err := newDB()
	if err != nil {
		log.Fatal("failed to initialze database: %v", err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			log.Error("failed to close db: %v", err)
		}
	}()

	http.Handle("/read", handler.Instrument(handler.RemoteRead(db)))
	http.Handle("/write", handler.Instrument(handler.RemoteWrite(db)))

	// must be last to allow all other instrumentation to be registered
	http.Handle("/metrics", promhttp.Handler())
	log.Info("http listening at %s", config.ListenAddr)
	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Error("listen http stopped: %v", err)
	}
	log.Info("server stopped")
}

func newDB() (*tsdb.DB, error) {
	return tsdb.Open(config.DataDir, nil, prometheus.DefaultRegisterer, &tsdb.Options{
		WALSegmentSize:    config.WALSegmentSize,
		RetentionDuration: uint64(config.RetentionDuration / time.Millisecond),
		BlockRanges:       tsdb.ExponentialBlockRanges(int64(time.Hour*2/time.Millisecond), 10, 3),
		NoLockfile:        config.NoLockFile,
	})
}
