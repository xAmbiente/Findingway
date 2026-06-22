// Package logger provides structured, leveled logging to both the console
// (with ANSI colour) and a rolling log file simultaneously.
package logger

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// Level represents log severity.
type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l Level) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	}
	return "?????"
}

// ANSI colour codes – only used for console output.
const (
	reset  = "\033[0m"
	bold   = "\033[1m"
	grey   = "\033[90m"
	cyan   = "\033[36m"
	green  = "\033[32m"
	yellow = "\033[33m"
	red    = "\033[31m"
	bRed   = "\033[1;31m"
)

func levelColor(l Level) string {
	switch l {
	case DEBUG:
		return grey
	case INFO:
		return cyan
	case WARN:
		return yellow
	case ERROR:
		return red
	case FATAL:
		return bRed
	}
	return reset
}

// Logger writes to console + file.
type Logger struct {
	mu      sync.Mutex
	file    *os.File
	noColor bool // set true when not a TTY (e.g. Docker)
}

var std *Logger

// Init creates the shared logger. Call once from main().
// logPath may be "" to skip file logging.
func Init(logPath string) error {
	l := &Logger{
		noColor: !isTTY(),
	}

	if logPath != "" {
		if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
			return fmt.Errorf("create log dir: %w", err)
		}
		f, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return fmt.Errorf("open log file: %w", err)
		}
		l.file = f
	}

	std = l
	return nil
}

// Close flushes and closes the log file.
func Close() {
	if std != nil && std.file != nil {
		std.file.Close()
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func (l *Logger) log(level Level, format string, args ...any) {
	msg := fmt.Sprintf(format, args...)
	ts := time.Now().Format("2006-01-02 15:04:05")

	// Plain line for the file (no colour codes)
	plain := fmt.Sprintf("%s  %-5s  %s\n", ts, level.String(), msg)

	// Coloured line for the console
	var console string
	if l.noColor {
		console = plain
	} else {
		console = fmt.Sprintf("%s%s%s  %s%s%s  %s\n",
			grey, ts, reset,
			levelColor(level)+bold, level.String(), reset,
			msg,
		)
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	// Choose stderr for WARN+ so it stands out in Docker logs
	var out io.Writer = os.Stdout
	if level >= WARN {
		out = os.Stderr
	}
	fmt.Fprint(out, console)

	if l.file != nil {
		fmt.Fprint(l.file, plain)
	}

	if level == FATAL {
		if l.file != nil {
			l.file.Close()
		}
		os.Exit(1)
	}
}

// Package-level helpers so callers just do logger.Info(...)

func Debug(format string, args ...any) { std.log(DEBUG, format, args...) }
func Info(format string, args ...any)  { std.log(INFO, format, args...) }
func Warn(format string, args ...any)  { std.log(WARN, format, args...) }
func Error(format string, args ...any) { std.log(ERROR, format, args...) }
func Fatal(format string, args ...any) { std.log(FATAL, format, args...) }

// Section prints a highlighted separator — useful for startup banners.
func Section(title string) {
	if std == nil {
		return
	}
	bar := "─────────────────────────────────────────"
	if std.noColor {
		std.mu.Lock()
		fmt.Fprintf(os.Stdout, "\n%s %s %s\n\n", bar, title, bar)
		if std.file != nil {
			fmt.Fprintf(std.file, "\n%s %s %s\n\n", bar, title, bar)
		}
		std.mu.Unlock()
	} else {
		std.mu.Lock()
		fmt.Fprintf(os.Stdout, "\n%s%s %s%s %s%s\n\n",
			cyan, bar, bold+title, reset+cyan, bar, reset)
		if std.file != nil {
			fmt.Fprintf(std.file, "\n%s %s %s\n\n", bar, title, bar)
		}
		std.mu.Unlock()
	}
}
