package log

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"io"
	"time"

	"github.com/natefinch/lumberjack"
)

var defaultLogger *zap.Logger

func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

func Panic(msg string, fields ...zap.Field) {
	defaultLogger.Panic(msg, fields...)
}

func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}

func Flush() {
	_ = defaultLogger.Sync()
}

func InitLogger(level zapcore.Level, ws ...io.Writer) {
	enc := zapcore.EncoderConfig{
		MessageKey:     "l",
		LevelKey:       "lv",
		TimeKey:        "time",
		NameKey:        "name",
		CallerKey:      "c",
		StacktraceKey:  "stack",
		LineEnding:     "\n",
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.RFC3339TimeEncoder,
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
	syncs := make([]zapcore.WriteSyncer, 0, len(ws))
	for _, w := range ws {
		sync := &zapcore.BufferedWriteSyncer{WS: zapcore.AddSync(w),
			FlushInterval: time.Second}
		syncs = append(syncs, sync)
	}
	out := zapcore.NewMultiWriteSyncer(syncs...)
	defaultLogger = zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(enc), out, level),
		zap.AddCaller(), zap.AddCallerSkip(1))
}

func RotateLogger(path, name string, sizeMb int) io.Writer {
	return &lumberjack.Logger{
		Filename:   fmt.Sprintf("%s/%s.log", path, name),
		MaxSize:    sizeMb,
		MaxAge:     14,
		MaxBackups: 50,
	}
}
