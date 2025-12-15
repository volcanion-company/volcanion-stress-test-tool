package logger

import (
	"io"
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.Logger

// LogConfig holds configuration for logging
type LogConfig struct {
	Level      string // debug, info, warn, error
	Format     string // json, console
	OutputPath string // file path or empty for stdout
	// Rotation settings (only applicable when OutputPath is set)
	MaxSizeMB  int  // max size in megabytes before rotation
	MaxBackups int  // max number of old log files to retain
	MaxAgeDays int  // max number of days to retain old log files
	Compress   bool // whether to compress rotated files
}

// DefaultLogConfig returns default logging configuration
func DefaultLogConfig() LogConfig {
	return LogConfig{
		Level:      "info",
		Format:     "json",
		OutputPath: "",
		MaxSizeMB:  100,
		MaxBackups: 5,
		MaxAgeDays: 30,
		Compress:   true,
	}
}

// Init initializes the global logger with structured JSON output
func Init(level string) error {
	config := DefaultLogConfig()
	config.Level = level
	return InitWithConfig(config)
}

// InitWithConfig initializes the global logger with full configuration
func InitWithConfig(config LogConfig) error {
	// Parse log level
	var zapLevel zapcore.Level
	switch config.Level {
	case "debug":
		zapLevel = zapcore.DebugLevel
	case "info":
		zapLevel = zapcore.InfoLevel
	case "warn":
		zapLevel = zapcore.WarnLevel
	case "error":
		zapLevel = zapcore.ErrorLevel
	default:
		zapLevel = zapcore.InfoLevel
	}

	// Configure encoder
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "timestamp",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "message",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.MillisDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var encoder zapcore.Encoder
	if config.Format == "console" {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	}

	// Configure output
	var writer zapcore.WriteSyncer
	if config.OutputPath != "" {
		// Ensure directory exists
		dir := filepath.Dir(config.OutputPath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// Setup lumberjack for log rotation
		lj := &lumberjack.Logger{
			Filename:   config.OutputPath,
			MaxSize:    config.MaxSizeMB,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAgeDays,
			Compress:   config.Compress,
			LocalTime:  true,
		}

		// Write to both file and stdout
		writer = zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(lj),
			zapcore.AddSync(os.Stdout),
		)
	} else {
		writer = zapcore.AddSync(os.Stdout)
	}

	// Create core
	core := zapcore.NewCore(encoder, writer, zapLevel)

	// Build logger with caller info
	Log = zap.New(core, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))

	return nil
}

// With creates a child logger with additional fields
func With(fields ...zap.Field) *zap.Logger {
	if Log == nil {
		return zap.NewNop()
	}
	return Log.With(fields...)
}

// WithRequestID creates a child logger with request ID field
func WithRequestID(requestID string) *zap.Logger {
	return With(zap.String("request_id", requestID))
}

// NewWriterAdapter creates an io.Writer that writes to the logger
func NewWriterAdapter(logger *zap.Logger, level zapcore.Level) io.Writer {
	return &writerAdapter{logger: logger, level: level}
}

type writerAdapter struct {
	logger *zap.Logger
	level  zapcore.Level
}

func (w *writerAdapter) Write(p []byte) (n int, err error) {
	msg := string(p)
	switch w.level {
	case zapcore.DebugLevel:
		w.logger.Debug(msg)
	case zapcore.InfoLevel:
		w.logger.Info(msg)
	case zapcore.WarnLevel:
		w.logger.Warn(msg)
	case zapcore.ErrorLevel:
		w.logger.Error(msg)
	case zapcore.DPanicLevel:
		w.logger.DPanic(msg)
	case zapcore.PanicLevel:
		w.logger.Panic(msg)
	case zapcore.FatalLevel:
		w.logger.Fatal(msg)
	case zapcore.InvalidLevel:
		w.logger.Info(msg)
	default:
		w.logger.Info(msg)
	}
	return len(p), nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Log != nil {
		_ = Log.Sync()
	}
}
