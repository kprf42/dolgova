package logger

import (
	"fmt"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger - обертка вокруг zap.Logger
type Logger struct {
	*zap.Logger
}

// LogConfig конфигурация для логгера
type LogConfig struct {
	Level      string // debug, info, warn, error, fatal
	OutputPath string // путь к файлу или "stdout" для вывода в консоль
	Format     string // json или console
}

// New создает новый экземпляр логгера с конфигурацией по умолчанию
func New() (*Logger, error) {
	return NewWithConfig(LogConfig{
		Level:      "info",
		OutputPath: "stdout",
		Format:     "console",
	})
}

// NewWithConfig создает новый экземпляр логгера с заданной конфигурацией
func NewWithConfig(config LogConfig) (*Logger, error) {
	// Настройка уровня логирования
	level := zap.NewAtomicLevel()
	err := level.UnmarshalText([]byte(config.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %v", err)
	}

	// Настройка вывода
	var outputPaths []string
	if config.OutputPath == "stdout" {
		outputPaths = []string{"stdout"}
	} else {
		outputPaths = []string{config.OutputPath}
	}

	// Настройка энкодера
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "ts",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		FunctionKey:    zapcore.OmitKey,
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zapConfig := zap.Config{
		Level:            level,
		Development:      false,
		Encoding:         config.Format,
		EncoderConfig:    encoderConfig,
		OutputPaths:      outputPaths,
		ErrorOutputPaths: []string{"stderr"},
	}

	zapLogger, err := zapConfig.Build(zap.AddCallerSkip(1))
	if err != nil {
		return nil, err
	}

	return &Logger{zapLogger}, nil
}

// Debug логирует сообщение с уровнем Debug
func (l *Logger) Debug(msg string, fields ...zap.Field) {
	l.Logger.Debug(msg, fields...)
}

// Info логирует сообщение с уровнем Info
func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.Logger.Info(msg, fields...)
}

// Warn логирует сообщение с уровнем Warn
func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.Logger.Warn(msg, fields...)
}

// Error логирует сообщение с уровнем Error
func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.Logger.Error(msg, fields...)
}

// Fatal логирует сообщение с уровнем Fatal и завершает программу
func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.Logger.Fatal(msg, fields...)
}

// Fatalf логирует форматированное сообщение с уровнем Fatal и завершает программу
func (l *Logger) Fatalf(format string, err error) {
	if err != nil {
		l.Logger.Fatal(format, zap.Error(err))
	}
	os.Exit(1)
}

// WithFields создает новый логгер с дополнительными полями
func (l *Logger) WithFields(fields ...zap.Field) *Logger {
	return &Logger{l.Logger.With(fields...)}
}

// Вспомогательные функции для создания полей
func String(key, val string) zap.Field {
	return zap.String(key, val)
}

func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

func Int64(key string, val int64) zap.Field {
	return zap.Int64(key, val)
}

func Float64(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}

func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

func Error(err error) zap.Field {
	return zap.Error(err)
}

func Any(key string, val interface{}) zap.Field {
	return zap.Any(key, val)
}

func Duration(key string, val float64) zap.Field {
	return zap.Float64(key, val)
}
