package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var log *zap.Logger

func Init(level, format, output, filePath string) error {
	var config zap.Config

	if format == "console" {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	} else {
		config = zap.NewProductionConfig()
	}

	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(zapLevel)

	if output == "file" && filePath != "" {
		config.OutputPaths = []string{filePath}
		config.ErrorOutputPaths = []string{filePath}
	}

	var err error
	log, err = config.Build()
	if err != nil {
		return err
	}

	return nil
}

func Get() *zap.Logger {
	if log == nil {

		log, _ = zap.NewDevelopment()
	}
	return log
}

func Sync() error {
	if log != nil {
		_ = log.Sync()
	}
	return nil
}

func Debug(msg string, fields ...zap.Field) {
	Get().Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	Get().Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	Get().Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	Get().Error(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	Get().Fatal(msg, fields...)
	os.Exit(1)
}

func Panic(msg string, fields ...zap.Field) {
	Get().Panic(msg, fields...)
}

func With(fields ...zap.Field) *zap.Logger {
	return Get().With(fields...)
}
