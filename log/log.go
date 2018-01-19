package log

import (
	"go.uber.org/zap"
)

var logger *zap.SugaredLogger

func init() {
	logger = zap.NewNop().Sugar()
}

func Initialize(debug bool) {
	options := []zap.Option{
		zap.AddCallerSkip(1),
	}

	var l *zap.Logger
	if debug {
		l, _ = zap.NewDevelopment(options...)
	} else {
		l, _ = zap.NewProduction(options...)
	}
	defer l.Sync()
	logger = l.Sugar()
}

func Info(args ...interface{}) {
	logger.Info(args...)
}
func Infow(msg string, args ...interface{}) {
	logger.Infow(msg, args...)
}
func Infof(msg string, args ...interface{}) {
	logger.Infof(msg, args...)
}

func Debug(args ...interface{}) {
	logger.Debug(args...)
}
func Debugw(msg string, args ...interface{}) {
	logger.Debugw(msg, args...)
}
func Debugf(msg string, args ...interface{}) {
	logger.Debugf(msg, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}
func Warnw(msg string, args ...interface{}) {
	logger.Warnw(msg, args...)
}
func Warnf(msg string, args ...interface{}) {
	logger.Warnf(msg, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}
func Errorw(msg string, args ...interface{}) {
	logger.Errorw(msg, args...)
}
func Errorf(msg string, args ...interface{}) {
	logger.Errorf(msg, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}
func Fatalw(msg string, args ...interface{}) {
	logger.Fatalw(msg, args...)
}
func Fatalf(msg string, args ...interface{}) {
	logger.Fatalf(msg, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}
func Panicw(msg string, args ...interface{}) {
	logger.Panicw(msg, args...)
}
func Panicf(msg string, args ...interface{}) {
	logger.Panicf(msg, args...)
}
