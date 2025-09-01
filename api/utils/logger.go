package utils

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"sync"
	"time"
)

var (
	appLogger *log.Logger
	logWriter io.WriteCloser
	initOnce  sync.Once
)

// Init initializes the application logger and ensures a log directory exists at the
// repository root. It prefers ./log relative to the current working directory,
// and falls back to ../log if needed, so it works when running from repo root
// or from the api directory.
func Init() (io.Writer, io.Closer, error) {
	var initErr error
	initOnce.Do(func() {
		// Resolve log directory: try ./log, else ../log
		candidates := []string{
			// Prefer repo root log dir (when running from api/)
			filepath.Join("..", "log"),
			// Fallback to current directory log (when running from repo root)
			filepath.Join(".", "log"),
		}

		var logDir string
		for _, dir := range candidates {
			if err := os.MkdirAll(dir, 0o755); err == nil {
				logDir = dir
				break
			}
		}
		if logDir == "" {
			initErr = fmt.Errorf("failed to create log directory in candidates: %v", candidates)
			return
		}

		// Daily log file name for simple rotation
		fileName := fmt.Sprintf("app-%s.log", time.Now().Format("2006-01-02"))
		filePath := filepath.Join(logDir, fileName)

		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			initErr = fmt.Errorf("failed to open log file: %w", err)
			return
		}

		logWriter = f
		appLogger = log.New(f, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	})

	return logWriter, logWriter, initErr
}

// Logger returns the initialized application logger. Call Init() first.
func Logger() *log.Logger {
	if appLogger == nil {
		// Best-effort fallback to stdout if Init wasn't called
		appLogger = log.New(os.Stdout, "", log.LstdFlags|log.LUTC|log.Lshortfile)
	}
	return appLogger
}

// Infof logs an informational message.
func Infof(format string, v ...any) {
	Logger().Printf("INFO: "+format, v...)
}

// Warnf logs a warning message.
func Warnf(format string, v ...any) {
	Logger().Printf("WARN: "+format, v...)
}

// Errorf logs an error message.
func Errorf(format string, v ...any) {
	Logger().Printf("ERROR: "+format, v...)
}
