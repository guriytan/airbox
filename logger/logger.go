package logger

import (
	"context"
	"fmt"

	"airbox/global"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var logger *zap.Logger

func InitializeLogger() {
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
		OutputPaths:      []string{"stdout", "/log/zap.log"},
		ErrorOutputPaths: []string{"stderr", "/log/zap.log"},
	}
	var err error
	logger, err = cfg.Build()
	if err != nil {
		panic(fmt.Sprintf("logger 初始化失败: %v", err))
	}
}

func GetLogger(ctx context.Context, funcName string) *zap.SugaredLogger {
	sugaredLogger := logger.Sugar().With(global.KeyFunction, funcName)
	return withField(ctx, sugaredLogger)
}

func withField(ctx context.Context, log *zap.SugaredLogger) *zap.SugaredLogger {
	if requestID, ok := ctx.Value(global.KeyRequestID).(string); ok {
		log = log.With(global.KeyRequestID, requestID)
	}
	if userID, ok := ctx.Value(global.KeyUserID).(string); ok {
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
