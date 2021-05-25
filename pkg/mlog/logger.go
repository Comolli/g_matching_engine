package mlog

import (
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjackv2 "gopkg.in/natefinch/lumberjack.v2"
)

var (
	Logger *zap.Logger
	Sugar  *zap.SugaredLogger
)

func TimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
}

func NewEncoderConfig() zapcore.EncoderConfig {
	return zapcore.EncoderConfig{
		// Keys can be anything except the empty string.
		TimeKey:        "Time",
		LevelKey:       "Level",
		NameKey:        "Name",
		CallerKey:      "Caller",
		MessageKey:     "Msg",
		StacktraceKey:  "Stack",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}
}
func Init() {
	zapConf := DefaultConfig(Console, zap.InfoLevel)
	w := zapcore.AddSync(&lumberjackv2.Logger{
		Filename:   zapConf.FileName,
		MaxSize:    1024, // megabytes
		MaxBackups: 10,
		MaxAge:     7, // days
	})

	var writeSyncer zapcore.WriteSyncer
	switch zapConf.stdout {
	case Console:
		writeSyncer = zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout))
	case File:
		writeSyncer = zapcore.NewMultiWriteSyncer(w)
	}
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(NewEncoderConfig()),
		writeSyncer,
		zapConf.Level,
	)
	Logger = zap.New(core, zap.AddCaller())
	Sugar = Logger.Sugar()
}
