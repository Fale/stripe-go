package stripe

import (
	"fmt"
	"io"
	"os"
	"regexp"
)

//
// Public constants
//

const (
	// LevelNull sets a logger to show no messages at all.
	LevelNull Level = 0

	// LevelError sets a logger to show error messages only.
	LevelError Level = 1

	// LevelWarn sets a logger to show warning messages or anything more
	// severe.
	LevelWarn Level = 2

	// LevelInfo sets a logger to show informational messages or anything more
	// severe.
	LevelInfo Level = 3

	// LevelDebug sets a logger to show informational messages or anything more
	// severe.
	LevelDebug Level = 4
)

//
// Public variables
//

// DefaultLeveledLogger is the default logger that the library will use to log
// errors, warnings, and informational messages.
//
// LeveledLoggerInterface is implemented by LeveledLogger, and one can be
// initialized at the desired level of logging.  LeveledLoggerInterface also
// provides out-of-the-box compatibility with a Logrus Logger, but may require
// a thin shim for use with other logging libraries that use less standard
// conventions like Zap.
//
// This Logger will be inherited by any backends created by default, but will
// be overridden if a backend is created with GetBackendWithConfig with a
// custom LeveledLogger set.
var DefaultLeveledLogger LeveledLoggerInterface = &LeveledLogger{
	Level: LevelError,
}

//
// Public types
//

// Level represents a logging level.
type Level uint32

// LeveledLogger is a leveled logger implementation.
//
// It prints warnings and errors to `os.Stderr` and other messages to
// `os.Stdout`.
type LeveledLogger struct {
	// Level is the minimum logging level that will be emitted by this logger.
	//
	// For example, a Level set to LevelWarn will emit warnings and errors, but
	// not informational or debug messages.
	//
	// Always set this with a constant like LevelWarn because the individual
	// values are not guaranteed to be stable.
	Level Level

	// Internal testing use only.
	stderrOverride io.Writer
	stdoutOverride io.Writer
}

var apiKeyRegex *regexp.Regexp

func init() {
	var err error
	apiKeyRegex, err = regexp.Compile("(sk_live|sk_test|rk_live|rk_test)_[a-zA-Z0-9]{2}[a-zA-Z0-9]{1,}")
	if err != nil {
		panic(err)
	}
}

// Redact the middle of a string as needed and replace it with an asterisk.
// Keep the given number of characters at the start and end of the string.
func redactMiddle(msg string, startLen int, endLen int) string {
	if len(msg) <= startLen+endLen {
		return msg
	}
	redactedSubstr := ""
	for i := 0; i < len(msg)-endLen-startLen; i++ {
		redactedSubstr += "*"
	}
	return msg[:startLen] + redactedSubstr + msg[len(msg)-endLen:]
}

func redact(msg string) string {
	return apiKeyRegex.ReplaceAllStringFunc(msg, func(str string) string {
		return redactMiddle(str, 10, 4)
	})
}

func (l *LeveledLogger) printLog(format string, tag string, level Level, output io.Writer, v ...interface{}) {
	if l.Level >= level {
		s := fmt.Sprintf(format, v...)
		fmt.Fprintf(output, "[%s] %s\n", tag, redact(s))
	}
}

// Debugf logs a debug message using Printf conventions.
func (l *LeveledLogger) Debugf(format string, v ...interface{}) {
	l.printLog(format, "DEBUG", LevelDebug, l.stdout(), v...)
}

// Errorf logs a warning message using Printf conventions.
func (l *LeveledLogger) Errorf(format string, v ...interface{}) {
	l.printLog(format, "ERROR", LevelError, l.stderr(), v...)
}

// Infof logs an informational message using Printf conventions.
func (l *LeveledLogger) Infof(format string, v ...interface{}) {
	l.printLog(format, "INFO", LevelInfo, l.stdout(), v...)
}

// Warnf logs a warning message using Printf conventions.
func (l *LeveledLogger) Warnf(format string, v ...interface{}) {
	l.printLog(format, "WARN", LevelWarn, l.stderr(), v...)
}

func (l *LeveledLogger) stderr() io.Writer {
	if l.stderrOverride != nil {
		return l.stderrOverride
	}

	return os.Stderr
}

func (l *LeveledLogger) stdout() io.Writer {
	if l.stdoutOverride != nil {
		return l.stdoutOverride
	}

	return os.Stdout
}

// LeveledLoggerInterface provides a basic leveled logging interface for
// printing debug, informational, warning, and error messages.
//
// It's implemented by LeveledLogger and also provides out-of-the-box
// compatibility with a Logrus Logger, but may require a thin shim for use with
// other logging libraries that you use less standard conventions like Zap.
type LeveledLoggerInterface interface {
	// Debugf logs a debug message using Printf conventions.
	Debugf(format string, v ...interface{})

	// Errorf logs a warning message using Printf conventions.
	Errorf(format string, v ...interface{})

	// Infof logs an informational message using Printf conventions.
	Infof(format string, v ...interface{})

	// Warnf logs a warning message using Printf conventions.
	Warnf(format string, v ...interface{})
}
