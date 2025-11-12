package logger

import (
	"os"
	"webrtc/p2p-server/pkg/config"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

var Log *zap.SugaredLogger

func InitLogger(cfg *config.Config) {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewJSONEncoder(encoderConfig),
		zapcore.NewMultiWriteSyncer(
			zapcore.AddSync(os.Stdout),
			zapcore.AddSync(&lumberjack.Logger{
				Filename:   cfg.Log.Path,
				MaxSize:    100, // MB
				MaxBackups: 7,
				MaxAge:     30, // days
				Compress:   true,
			}),
		),
		zap.InfoLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.Development())
	Log = logger.Sugar()
}
