package logger

import (
	"fmt"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
	"log"
	"os"
	"strings"
	"time"
)

var sugaredLogger *zap.SugaredLogger

func Init(to, level, projectName string) {
	var cores []zapcore.Core

	var lvl zapcore.Level
	err := lvl.UnmarshalText([]byte(level))
	if err != nil {
		log.Println("invalid level:", level)
		return
	}

	// file
	fileEncoder := getEncoder()
	writer := zapcore.AddSync(&lumberjack.Logger{
		Filename:   fmt.Sprintf("/data/%s/logs/%s.log", projectName, level),
		MaxSize:    200, // megabytes
		MaxBackups: 3,
		MaxAge:     30, // days
	})
	logTo := strings.ToUpper(to)
	if logTo != "" || strings.Contains(logTo, "CONSOLE") {
		cores = append(cores, zapcore.NewCore(getEncoder(), zapcore.Lock(os.Stdout), lvl))
	}
	if logTo != "" && strings.Contains(logTo, "FILE") {
		cores = append(cores, zapcore.NewCore(fileEncoder, writer, lvl))
	}
	core := zapcore.NewTee(cores...)

	sugaredLogger = zap.New(core).Sugar()
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format("2006-01-02 15:04:05.000"))
	}

	return zapcore.NewConsoleEncoder(encoderConfig)
}

func Infow(msg string, args ...interface{}) {
	sugaredLogger.Infow(msg, args...)
}
func Debug(args ...interface{}) {
	sugaredLogger.Debug(args...)
}
func Debugf(msg string, args ...interface{}) {
	sugaredLogger.Debugf(msg, args...)
}
func Info(args ...interface{}) {
	sugaredLogger.Info(args...)
}
func Infof(msg string, args ...interface{}) {
	sugaredLogger.Infof(msg, args...)
}
func Error(args ...interface{}) {
	sugaredLogger.Error(args...)
}
func Errorf(msg string, args ...interface{}) {
	sugaredLogger.Errorf(msg, args...)
}
func Warn(args ...interface{}) {
	sugaredLogger.Warn(args...)
}
