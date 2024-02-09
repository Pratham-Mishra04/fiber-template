package initializers

import (
	"log"
	"os"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

const (
	DebugLevel = zapcore.DebugLevel // DebugLevel logs are disabled in production.
	InfoLevel  = zapcore.InfoLevel  // InfoLevel logs are Debug priority logs.
	WarnLevel  = zapcore.WarnLevel  // WarnLevel logs are more important than Info, but don't need individual human review.
	ErrorLevel = zapcore.ErrorLevel // ErrorLevel logs are high-priority.
	PanicLevel = zapcore.PanicLevel // PanicLevel logs a message, then panics.
	FatalLevel = zapcore.FatalLevel // FatalLevel logs a message, then calls os.Exit(1).
)

const (
	DebugLogFilePath = "logs/debug.log"
	InfoLogFilePath  = "logs/info.log"
	WarnLogFilePath  = "logs/warn.log"
	ErrorLogFilePath = "logs/error.log"
	PanicLogFilePath = "logs/panic.log"
	FatalLogFilePath = "logs/fatal.log"
)

var logFiles struct {
	DebugLogFile *os.File
	InfoLogFile  *os.File
	WarnLogFile  *os.File
	ErrorLogFile *os.File
	PanicLogFile *os.File
	FatalLogFile *os.File
}

func AddLogger() {
	openLogFiles()

	debugCore := newCore(logFiles.DebugLogFile, DebugLevel)
	infoCore := newCore(logFiles.InfoLogFile, InfoLevel)
	warnCore := newCore(logFiles.WarnLogFile, WarnLevel)
	errorCore := newCore(logFiles.ErrorLogFile, ErrorLevel)
	panicCore := newCore(logFiles.PanicLogFile, PanicLevel)
	fatalCore := newCore(logFiles.FatalLogFile, FatalLevel)

	Logger = zap.New(zapcore.NewTee(debugCore, infoCore, warnCore, errorCore, panicCore, fatalCore)).Sugar()
}

func newCore(LogFile *os.File, LoggerLevel zapcore.Level) zapcore.Core {
	encoderCfg := zap.NewProductionEncoderConfig()
	encoderCfg.EncodeTime = zapcore.TimeEncoderOfLayout(time.RFC1123)
	fileCore := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderCfg),
		zapcore.AddSync(LogFile),
		LoggerLevel,
	)
	return fileCore
}

func openLogFiles() {
	logFiles.DebugLogFile = openFile(DebugLogFilePath)
	logFiles.InfoLogFile = openFile(InfoLogFilePath)
	logFiles.WarnLogFile = openFile(WarnLogFilePath)
	logFiles.ErrorLogFile = openFile(ErrorLogFilePath)
	logFiles.PanicLogFile = openFile(PanicLogFilePath)
	logFiles.FatalLogFile = openFile(FatalLogFilePath)
}

func openFile(LogFilePath string) *os.File {
	logFile, err := os.OpenFile(LogFilePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("Failed to open log file: " + err.Error())
	}
	return logFile
}

func LoggerCleanUp() {
	logFiles.DebugLogFile.Close()
	logFiles.InfoLogFile.Close()
	logFiles.WarnLogFile.Close()
	logFiles.ErrorLogFile.Close()
	logFiles.PanicLogFile.Close()
	logFiles.FatalLogFile.Close()
	Logger.Sync()
}
