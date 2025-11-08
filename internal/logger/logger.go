package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func New(env string) *zap.SugaredLogger {
	level := zapcore.InfoLevel
	if env != "prod" {
		level = zapcore.DebugLevel
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(cfg),
		zapcore.AddSync(os.Stdout),
		level,
	)
	return zap.New(core).Sugar()
}
