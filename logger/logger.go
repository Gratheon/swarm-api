package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// ANSI color codes
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorGrey   = "\033[90m"
)

type UserCycleFormatter struct{}

func (f *UserCycleFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	// Time
	timestamp := entry.Time.Format("15:04:05")
	timestampColored := colorBlue + timestamp + colorReset

	// Level
	var levelColored string
	switch entry.Level {
	case logrus.InfoLevel:
		levelColored = colorGreen + "info" + colorReset
	case logrus.WarnLevel:
		levelColored = colorYellow + "warn" + colorReset
	case logrus.ErrorLevel, logrus.FatalLevel, logrus.PanicLevel:
		levelColored = colorRed + "error" + colorReset
	case logrus.DebugLevel:
		levelColored = colorGrey + "debug" + colorReset
	default:
		levelColored = string(entry.Level)
	}

	// Message
	msg := entry.Message
	if entry.Level == logrus.DebugLevel {
		msg = colorGrey + msg + colorReset
	}

	// Meta fields
	metaStr := ""
	for k, v := range entry.Data {
		metaStr += fmt.Sprintf(" %s%s%s=%s%v%s", colorPurple, k, colorReset, colorPurple, v, colorReset)
	}
	metaStr = strings.TrimSpace(metaStr)

	logLine := fmt.Sprintf("%s [%s]: %s %s\n", timestampColored, levelColored, msg, metaStr)
	return []byte(logLine), nil
}

func InitLogging() *logrus.Logger {
	logrusInstance := logrus.New()
	logrusInstance.SetOutput(os.Stdout)
	logrusInstance.SetLevel(logrus.InfoLevel)
	logrusInstance.SetFormatter(&UserCycleFormatter{})
	return logrusInstance
}

var logger = InitLogging()

func Info(message string, meta ...interface{}) {
	if len(meta) > 0 {
		logger.WithFields(toFields(meta...)).Info(message)
	} else {
		logger.Info(message)
	}
}

func Warn(message string, meta ...interface{}) {
	if len(meta) > 0 {
		logger.WithFields(toFields(meta...)).Warn(message)
	} else {
		logger.Warn(message)
	}
}

func Debug(message string, meta ...interface{}) {
	if len(meta) > 0 {
		logger.WithFields(toFields(meta...)).Debug(message)
	} else {
		logger.Debug(message)
	}
}

func Error(message string, meta ...interface{}) {
	if len(meta) > 0 {
		logger.WithFields(toFields(meta...)).Error(message)
	} else {
		logger.Error(message)
	}
}

func Fatal(message string, meta ...interface{}) {
	if len(meta) > 0 {
		logger.WithFields(toFields(meta...)).Fatal(message)
	} else {
		logger.Fatal(message)
	}
}

func toFields(meta ...interface{}) logrus.Fields {
	fields := logrus.Fields{}
	for i, m := range meta {
		fields["meta_"+string(rune(i+'0'))] = m
	}
	return fields
}
