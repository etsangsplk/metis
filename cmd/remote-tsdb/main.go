package main

import (
	"flag"
	"fmt"
	"math"
	"net/http"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/blockloop/remote-tsdb/handler"
	"github.com/blockloop/remote-tsdb/log"
	"github.com/prometheus/tsdb"
)

var config struct {
	ListenAddr        string
	DataDir           string
	RetentionDuration time.Duration
	NoLockFile        bool
	WALSegmentSize    int

	LogLevel string
}

// Init initializes configuration
func Init() error {
	flag.StringVar(&config.LogLevel, "log-level", "info", "log level info or error")
	flag.StringVar(&config.ListenAddr, "listen-addr", ":8080", "web address to listen for RemoteRead and RemoteWrite")
	flag.StringVar(&config.DataDir, "data-dir", "./data", "filesystem path to store tsdb data")
	flag.DurationVar(&config.RetentionDuration, "retention-duration", time.Hour*48, "tsdb retention time")
	flag.BoolVar(&config.NoLockFile, "no-lockfile", false, "disable the use of the TSDB lockfile")

	var wals string
	flag.StringVar(&wals, "wal-segment-size", "128MB", "Write Ahead Log segment size")
	flag.Parse()

	s, err := bytefmt.ToBytes(wals)
	if err != nil {
		return fmt.Errorf("wal-segment-size is not a valid format. Expected format: 128MB, 1GB. Got %q\n", wals)
	}
	if s > math.MaxInt32 {
		return fmt.Errorf("wal-segment-size must be less than %s. Got %q\n", bytefmt.ByteSize(math.MaxInt32), wals)
	}
	config.WALSegmentSize = int(s)

	if err := log.SetLevelString(config.LogLevel); err != nil {
		log.Fatal("log level %q is not understood: %v", config.LogLevel, err)
	}
	return nil
}

func main() {
	if err := Init(); err != nil {
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

	http.Handle("/read", handler.LogMW(handler.RemoteRead(db)))
	http.Handle("/write", handler.LogMW(handler.RemoteWrite(db)))

	log.Info("http listening at %s", config.ListenAddr)
	if err := http.ListenAndServe(config.ListenAddr, nil); err != nil {
		log.Error("listen http stopped: %v", err)
	}
	log.Info("server stopped")
}

func newDB() (*tsdb.DB, error) {
	return tsdb.Open(config.DataDir, nil, nil, &tsdb.Options{
		WALSegmentSize:    config.WALSegmentSize,
		RetentionDuration: uint64(config.RetentionDuration / time.Millisecond),
		BlockRanges:       tsdb.ExponentialBlockRanges(int64(time.Hour*2/time.Millisecond), 10, 3),
		NoLockfile:        config.NoLockFile,
	})
}
