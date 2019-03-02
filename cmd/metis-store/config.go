package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"time"

	"code.cloudfoundry.org/bytefmt"
	"github.com/digitalocean/metis/log"
)

var config struct {
	ListenAddr        string
	DataDir           string
	RetentionDuration time.Duration
	NoLockFile        bool
	WALSegmentSize    int

	LogLevel string
}

// InitConfig initializes configuration
func InitConfig() error {
	flag.StringVar(&config.LogLevel, "log-level", "info", "log level info or error")
	flag.StringVar(&config.ListenAddr, "listen-addr", ":8080", "web address to listen for RemoteRead and RemoteWrite")
	flag.StringVar(&config.DataDir, "data-dir", "./data", "filesystem path to store tsdb data")
	flag.DurationVar(&config.RetentionDuration, "retention-duration", time.Hour*48, "tsdb retention time")
	flag.BoolVar(&config.NoLockFile, "no-lockfile", false, "disable the use of the TSDB lockfile")

	var version bool
	flag.BoolVar(&version, "version", false, "show version information")

	var wals string
	flag.StringVar(&wals, "wal-segment-size", "128MB", "Write Ahead Log segment size")
	flag.Parse()

	if version {
		showVersion()
		os.Exit(0)
	}

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
