package logger

import (
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	Logger *zap.Logger
	once   sync.Once
	Log    = Sugar()
)

type Config struct {
	LogLevel string // "debug", "info", "warn", "error"
	DevMode  bool   // If true, uses development config with pretty printing
}

// Initialize sets up the logger with the given configuration
func Initialize(config Config) error {
	var err error
	once.Do(func() {
		// Set the log level
		level, err := zapcore.ParseLevel(config.LogLevel)
		if err != nil {
			level = zapcore.InfoLevel
		}

		// Configure encoder
		encoderConfig := zap.NewProductionEncoderConfig()
		encoderConfig.TimeKey = "timestamp"
		encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
		encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder

		// Create core
		core := zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)

		// Create logger
		Logger = zap.New(core)
		if config.DevMode {
			Logger = Logger.WithOptions(zap.Development())
		}
	})
	return err
}

// GetLogger returns the configured logger instance
func GetLogger() *zap.Logger {
	if Logger == nil {
		// If logger hasn't been initialized, create a default one
		Initialize(Config{
			LogLevel: "info",
			DevMode:  false,
		})
	}
	return Logger
}

// Sugar returns sugared logger
func Sugar() *zap.SugaredLogger {
	return GetLogger().Sugar()
}

func Info(args ...interface{}) {
	Log.Info(args...)
}

func Fatalf(template string, args ...interface{}) {
	Log.Fatalf(template, args...)
}
