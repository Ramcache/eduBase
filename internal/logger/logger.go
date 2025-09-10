package logger

import "go.uber.org/zap"

var Log *zap.Logger

func Init(env string) {
	var l *zap.Logger
	if env == "prod" {
		l, _ = zap.NewProduction()
	} else {
		l, _ = zap.NewDevelopment()
	}
	Log = l
}

func Err(err error) zap.Field       { return zap.Error(err) }
func Str(k, v string) zap.Field     { return zap.String(k, v) }
func Int(k string, v int) zap.Field { return zap.Int(k, v) }
