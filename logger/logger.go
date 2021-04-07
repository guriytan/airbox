package logger

import (
	"context"
	"fmt"
	"os"

	"airbox/global"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func InitializeLogger() {
	if _, err := os.Create(global.LogPath); err != nil {
		panic(fmt.Sprintf("log 初始化失败: %v", err))
	}
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "air_box",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "trace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.RFC3339TimeEncoder,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:       []string{"stdout", global.LogPath},
		ErrorOutputPaths:  []string{"stderr", global.LogPath},
		DisableStacktrace: true,
	}
	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("logger 初始化失败: %v", err))
	}
}

func GetLogger(ctx context.Context, funcName string) *customLogger {
	sugaredLogger := logger.Sugar().With(global.KeyFunction, funcName)
	return &customLogger{SugaredLogger: withField(ctx, sugaredLogger)}
}

type customLogger struct {
	*zap.SugaredLogger
}

func (log *customLogger) WithError(err error) *zap.SugaredLogger {
	return log.With("err", err)
}

func withField(ctx context.Context, log *zap.SugaredLogger) *zap.SugaredLogger {
	if requestID, ok := ctx.Value(global.KeyRequestID).(string); ok {
		log = log.With(global.KeyRequestID, requestID)
	}
	if userID, ok := ctx.Value(global.KeyUserID).(int64); ok {
		log = log.With(global.KeyUserID, userID)
	}
	if ip, ok := ctx.Value(global.KeyIP).(string); ok {
		log = log.With(global.KeyIP, ip)
	}
	return log
}

func CloseLogger() {
	_ = logger.Sync()
}
