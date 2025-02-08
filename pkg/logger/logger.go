package logger

import (
	"fmt"
	"os"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	once sync.Once
	log  *zap.SugaredLogger
)

type Config struct {
	LogLevel string // "debug", "info", "warn", "error"
	DevMode  bool   // If true, uses development config with pretty printing
	File     string
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

		// Configure lumberjack for log rotation
		fmt.Printf("Configuring logging with file: '%v'\n", config.File)

		logWriter := &lumberjack.Logger{
			Filename:   config.File, // Log file name
			MaxSize:    5,           // Max size in MB before rotation
			MaxBackups: 3,           // Number of old log files to retain
			MaxAge:     20,          // Max age in days to retain old logs
			Compress:   true,        // Compress old logs (gzip)
		}

		consoleEncoder := zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig())
		fileEncoder := zapcore.NewJSONEncoder(zap.NewProductionEncoderConfig())

		// Create core
		core := zapcore.NewTee(
			zapcore.NewCore(consoleEncoder, zapcore.AddSync(os.Stdout), zapcore.DebugLevel),
			zapcore.NewCore(fileEncoder, zapcore.AddSync(logWriter), zapcore.InfoLevel),
		)

		zapcore.NewCore(
			zapcore.NewJSONEncoder(encoderConfig),
			zapcore.AddSync(os.Stdout),
			level,
		)

		// Create logger
		_log := zap.New(core)
		if config.DevMode {
			_log = _log.WithOptions(zap.Development())
		}
		log = _log.Sugar()
	})
	return err
}

func Info(args ...interface{}) {
	log.Info(args...)
}

func Infof(template string, args ...interface{}) {
	log.Infof(template, args...)
}

func Debug(args ...interface{}) {
	log.Debug(args...)
}

func Debugf(template string, args ...interface{}) {
	log.Debugf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}

func Warningf(template string, args ...interface{}) {
	log.Fatalf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	log.Errorf(template, args...)
}
