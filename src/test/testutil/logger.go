package testutil

import (
	"io"

	"github.com/rs/zerolog"
	logger "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Logger"
	config "gitlab.com/maplesense1/mpt.mqtt_server/src/production/MQT.Config"
)

// NewTestLogger creates a logger for tests. If w is nil, uses stdout.
func NewTestLogger(w io.Writer) *logger.Logger {
	cfg := &config.LoggingConfig{
		Level:        "debug",
		Format:       "json",
		Output:       "stdout",
		EnableCaller: false,
	}
	l := logger.NewLogger(cfg)
	if w != nil {
		zl := zerolog.New(w).With().Timestamp().Logger()
		l = &logger.Logger{Logger: &zl}
	}
	return l
}

// NewDiscardLogger creates a logger that discards all output (for quiet tests)
func NewDiscardLogger() *logger.Logger {
	return NewTestLogger(io.Discard)
}
