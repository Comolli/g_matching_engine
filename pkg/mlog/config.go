package mlog

import (
	"go.uber.org/zap/zapcore"
)

type stdoutType int

const (
	Console stdoutType = iota
	File    stdoutType = iota
)

const (
	// DebugLevel logs are typically voluminous, and are usually disabled in
	// production.
	DebugLevel = zapcore.DebugLevel
	// InfoLevel is the default logging priority.
	InfoLevel = zapcore.InfoLevel
	// WarnLevel logs are more important than Info, but don't need individual
	// human review.
	WarnLevel = zapcore.WarnLevel
	// ErrorLevel logs are high-priority. If an application is running smoothly,
	// it shouldn't generate any error-level logs.
	ErrorLevel = zapcore.ErrorLevel
	// DPanicLevel logs are particularly important errors. In development the
	// logger panics after writing the message.
	DPanicLevel = zapcore.DPanicLevel
	// PanicLevel logs a message, then panics.
	PanicLevel = zapcore.PanicLevel
	// FatalLevel logs a message, then calls os.Exit(1).
	FatalLevel = zapcore.FatalLevel
)

type ZapConfig struct {
	stdout   stdoutType
	Level    zapcore.Level
	FileName string
}

func DefaultConfig(stdoutType stdoutType, Level zapcore.Level) *ZapConfig {
	var fileName string
	switch stdoutType {
	case Console:
	case File:
		fileName = "./g_m_log.log"
	}
	return &ZapConfig{
		stdout:   stdoutType,
		Level:    Level,
		FileName: fileName,
	}
}
