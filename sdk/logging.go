package sdk

import (
	"github.com/sirupsen/logrus"
)

// LogLevel 日志级别（参考 logrus 的日志级别）
type LogLevel string

const (
	LogLevelTrace LogLevel = "Trace"
	LogLevelDebug LogLevel = "Debug"
	LogLevelInfo  LogLevel = "Info"
	LogLevelWarn  LogLevel = "Warn"
	LogLevelError LogLevel = "Error"
	LogLevelFatal LogLevel = "Fatal"
	LogLevelPanic LogLevel = "Panic"
)

// stringToLogrusLevel 将字符串转换为 logrus 级别
func stringToLogrusLevel(levelStr string) logrus.Level {
	switch levelStr {
	case "Trace":
		return logrus.TraceLevel
	case "Debug":
		return logrus.DebugLevel
	case "Info":
		return logrus.InfoLevel
	case "Warn":
		return logrus.WarnLevel
	case "Error":
		return logrus.ErrorLevel
	case "Fatal":
		return logrus.FatalLevel
	case "Panic":
		return logrus.PanicLevel
	default:
		return logrus.InfoLevel
	}
}

// logrusLevelToLogLevel 将 logrus 级别转换为 SDK 日志级别
func logrusLevelToLogLevel(level logrus.Level) LogLevel {
	switch level {
	case logrus.TraceLevel:
		return LogLevelTrace
	case logrus.DebugLevel:
		return LogLevelDebug
	case logrus.InfoLevel:
		return LogLevelInfo
	case logrus.WarnLevel:
		return LogLevelWarn
	case logrus.ErrorLevel:
		return LogLevelError
	case logrus.FatalLevel:
		return LogLevelFatal
	case logrus.PanicLevel:
		return LogLevelPanic
	default:
		return LogLevelInfo
	}
}

// shouldReportLog 判断是否应该上报日志（只有大于等于配置的级别才上报）
func shouldReportLog(logLevel LogLevel, minLevel LogLevel) bool {
	levels := map[LogLevel]int{
		LogLevelTrace: 0,
		LogLevelDebug: 1,
		LogLevelInfo:  2,
		LogLevelWarn:  3,
		LogLevelError: 4,
		LogLevelFatal: 5,
		LogLevelPanic: 6,
	}
	
	logLevelValue, ok1 := levels[logLevel]
	minLevelValue, ok2 := levels[minLevel]
	
	if !ok1 || !ok2 {
		return true // 默认上报
	}
	
	return logLevelValue >= minLevelValue
}
