package logger

import (
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
)

// Logger wraps zerolog.Logger with additional functionality
type Logger struct {
	*zerolog.Logger
}

// NewLogger creates a new logger based on configuration
func NewLogger(cfg *config.LoggingConfig) *Logger {
	// Set log level
	level, err := zerolog.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		level = zerolog.InfoLevel
	}
	zerolog.SetGlobalLevel(level)

	// Configure output
	var output zerolog.ConsoleWriter
	if cfg.Format == "json" {
		// JSON output
		if cfg.Output == "stderr" {
			log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
		} else {
			log.Logger = zerolog.New(os.Stdout).With().Timestamp().Logger()
		}
	} else {
		// Console output
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		// Note: WithCaller is not available on ConsoleWriter in this version
		// Caller information will be handled separately for JSON format

		log.Logger = log.Output(output).With().Timestamp().Logger()
	}

	// Set caller information if enabled
	if cfg.EnableCaller && cfg.Format == "json" {
		log.Logger = log.Logger.With().Caller().Logger()
	}

	return &Logger{&log.Logger}
}

// WithField adds a field to the logger
func (l *Logger) WithField(key string, value interface{}) *Logger {
	logger := l.Logger.With().Interface(key, value).Logger()
	return &Logger{&logger}
}

// WithFields adds multiple fields to the logger
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	context := l.Logger.With()
	for key, value := range fields {
		context = context.Interface(key, value)
	}
	logger := context.Logger()
	return &Logger{&logger}
}

// WithError adds an error to the logger
func (l *Logger) WithError(err error) *Logger {
	logger := l.Logger.With().Err(err).Logger()
	return &Logger{&logger}
}

// WithRequestID adds a request ID to the logger
func (l *Logger) WithRequestID(requestID string) *Logger {
	logger := l.Logger.With().Str("request_id", requestID).Logger()
	return &Logger{&logger}
}

// WithService adds a service name to the logger
func (l *Logger) WithService(service string) *Logger {
	logger := l.Logger.With().Str("service", service).Logger()
	return &Logger{&logger}
}

// WithComponent adds a component name to the logger
func (l *Logger) WithComponent(component string) *Logger {
	logger := l.Logger.With().Str("component", component).Logger()
	return &Logger{&logger}
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.Logger.Fatal().Msg(msg)
}

// FatalWithError logs a fatal message with error and exits
func (l *Logger) FatalWithError(err error, msg string) {
	l.Logger.Fatal().Err(err).Msg(msg)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.Logger.Error().Msg(msg)
}

// ErrorWithError logs an error message with error
func (l *Logger) ErrorWithError(err error, msg string) {
	l.Logger.Error().Err(err).Msg(msg)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.Logger.Warn().Msg(msg)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.Logger.Info().Msg(msg)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.Logger.Debug().Msg(msg)
}

// Trace logs a trace message
func (l *Logger) Trace(msg string) {
	l.Logger.Trace().Msg(msg)
}

// GetGlobalLogger returns the global logger instance
func GetGlobalLogger() *Logger {
	return &Logger{&log.Logger}
}

// SetGlobalLogger sets the global logger instance
func SetGlobalLogger(logger *Logger) {
	log.Logger = *logger.Logger
}
