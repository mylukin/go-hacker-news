package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

var (
	// Debug模式标志
	debugMode bool
	// 日志实例
	infoLogger  *log.Logger
	debugLogger *log.Logger
	warnLogger  *log.Logger
	errorLogger *log.Logger
)

// 初始化日志系统
func Init(debug bool) {
	debugMode = debug

	// 通用日志前缀格式
	logFlags := log.Ldate | log.Ltime

	// 创建不同级别的日志记录器
	infoLogger = log.New(os.Stdout, "[INFO] ", logFlags)
	debugLogger = log.New(os.Stdout, "[DEBUG] ", logFlags)
	warnLogger = log.New(os.Stdout, "[WARN] ", logFlags)
	errorLogger = log.New(os.Stderr, "[ERROR] ", logFlags)
}

// SetDebugMode 动态设置调试模式
func SetDebugMode(debug bool) {
	debugMode = debug
	if debug {
		Debug("调试日志已启用")
	}
}

// Info 记录重要的信息性日志，无论是否在debug模式都会显示
func Info(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	infoLogger.Println(msg)
}

// Debug 记录调试信息，仅在debug模式下显示
func Debug(format string, args ...interface{}) {
	if debugMode {
		msg := fmt.Sprintf(format, args...)
		debugLogger.Println(msg)
	}
}

// Warn 记录警告信息，无论是否在debug模式都会显示
func Warn(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	warnLogger.Println(msg)
}

// Error 记录错误信息，无论是否在debug模式都会显示
func Error(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	errorLogger.Println(msg)
}

// Timer 用于计时和记录操作耗时
type Timer struct {
	name      string
	startTime time.Time
}

// NewTimer 创建一个新的计时器
func NewTimer(name string) *Timer {
	t := &Timer{
		name:      name,
		startTime: time.Now(),
	}
	if debugMode {
		Debug("开始操作: %s", name)
	}
	return t
}

// Stop 停止计时并记录耗时
func (t *Timer) Stop() {
	if debugMode {
		duration := time.Since(t.startTime)
		Debug("完成操作: %s (耗时: %v)", t.name, duration)
	}
}
