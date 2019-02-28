package log

import (
	"errors"
	"fmt"
	"log"
	"log/syslog"
	"os"
)

// Level is a log level such a Info or Error
type Level int

const (
	syslogFlags = log.Lshortfile
	normalFlags = log.LUTC | log.Ldate | log.Ltime | log.Lshortfile

	// LevelInfo enables info logging
	LevelInfo Level = iota
	// LevelError enables error logging
	LevelError Level = iota
)

var (
	// ErrBadLogLevel means the log level provided to SetLevelString is not understood
	ErrBadLogLevel = errors.New("bad log level")

	infolog = log.New(os.Stdout, "", normalFlags)
	errlog  = log.New(os.Stderr, "", normalFlags)

	level = LevelError
)

// SetLevel sets the log level
func SetLevel(l Level) {
	level = l
}

// SetLevelString sets the log level from the provided string
// if the string cannot be understood then panic
func SetLevelString(s string) error {
	switch s {
	case "info", "INFO":
		SetLevel(LevelInfo)
	case "error", "ERROR":
		SetLevel(LevelError)
	default:
		return ErrBadLogLevel
	}
	return nil
}

// InitSyslog initializes logging to syslog
func InitSyslog() (err error) {
	dl, err := syslog.NewLogger(syslog.LOG_NOTICE, syslogFlags)
	if err != nil {
		return fmt.Errorf("InitSyslog failed to initialize info logger: %+v", err)
	}
	infolog = dl

	el, err := syslog.NewLogger(syslog.LOG_ERR, syslogFlags)
	if err != nil {
		return fmt.Errorf("InitSyslog failed to initialize error logger: %+v", err)
	}
	errlog = el

	return nil
}

// Info prints a info message. If syslog is enabled then LOG_NOTICE is used
func Info(msg string, params ...interface{}) {
	msg = "INFO " + msg
	if level > LevelInfo {
		return
	}

	if err := infolog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Error prints an error message. If syslog is enabled then LOG_ERR is used
func Error(msg string, params ...interface{}) {
	msg = "ERROR " + msg
	if err := errlog.Output(2, fmt.Sprintf(msg, params...)); err != nil {
		fmt.Fprintf(os.Stderr, "ERROR writing log output: %+v", err)
	}
}

// Fatal logs Error and exits 1
func Fatal(msg string, params ...interface{}) {
	Error(msg, params...)
	os.Exit(1)
}
