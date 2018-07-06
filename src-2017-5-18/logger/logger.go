
//提供一个分等级的日志系统，建议直接使用全局的对象，而不是另外New一个
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"runtime/debug"
)

func init() {
	globalLogger = New(os.Stdout, "", log.LstdFlags, INFO)
}

var globalLogger *Logger

const (
	DEBUG = iota
	INFO
	WARNING
	ERROR
	FATAL
	NONE
)

var levelNames = []string{
	"DEBUG",
	"INFO",
	"WARNING",
	"ERROR",
	"FATAL",
	"NONE",
}

var levelPrefixes []string

func init() {
	levelPrefixes = make([]string, len(levelNames))
	for i, name := range levelNames {
		levelPrefixes[i] = name + ": "
	}
}

func Debug(format string, args ...interface{}) {
	globalLogger.Output(DEBUG, format, args...)
}

func Info(format string, args ...interface{}) {
	globalLogger.Output(INFO, format, args...)
}

func Warning(format string, args ...interface{}) {
	globalLogger.Output(WARNING, format, args...)
}

func Error(format string, args ...interface{}) {
	globalLogger.Output(ERROR, format, args...)
}

func Fatal(format string, args ...interface{}) {
	globalLogger.Output(FATAL, format, args...)
	debug.PrintStack()
	os.Exit(1)
}

func SetLogger(logger *Logger) {
	globalLogger = logger
}

type Logger struct {
	logger *log.Logger
	level  int
}

func New(out io.Writer, prefix string, flag, level int) *Logger {
	return &Logger{logger: log.New(out, prefix, flag), level: level}
}

func (logger *Logger) Debug(format string, args ...interface{}) {
	logger.Output(DEBUG, format, args...)
}

func (logger *Logger) Info(format string, args ...interface{}) {
	logger.Output(INFO, format, args...)
}

func (logger *Logger) Warning(format string, args ...interface{}) {
	logger.Output(WARNING, format, args...)
}

func (logger *Logger) Error(format string, args ...interface{}) {
	logger.Output(ERROR, format, args...)
}

func (logger *Logger) Fatal(format string, args ...interface{}) {
	logger.Output(FATAL, format, args...)
	debug.PrintStack()
	os.Exit(1)
}

// 如果对象包含需要加密的信息（例如密码），请实现Redactor接口
type Redactor interface {
	// 返回一个去处掉敏感信息的示例
	Redacted() interface{}
}

// Redact 返回跟字符串等长的“＊”。
func Redact(s string) string {
	return strings.Repeat("*", len(s))
}

func (logger *Logger) Output(level int, format string, args ...interface{}) {
	if logger.level > level {
		return
	}
	redactedArgs := make([]interface{}, len(args))
	for i, arg := range args {
		if redactor, ok := arg.(Redactor); ok {
			redactedArgs[i] = redactor.Redacted()
		} else {
			redactedArgs[i] = arg
		}
	}
	logger.logger.Output(3, levelPrefixes[level]+fmt.Sprintf(format, redactedArgs...))
}

func (logger *Logger) SetFlags(flag int) {
	logger.logger.SetFlags(flag)
}

func (logger *Logger) SetPrefix(prefix string) {
	logger.logger.SetPrefix(prefix)
}

func (logger *Logger) SetLevel(level int) {
	logger.level = level
}

func LogNameToLogLevel(name string) int {
	s := strings.ToUpper(name)
	for i, level := range levelNames {
		if level == s {
			return i
		}
	}
	panic(fmt.Errorf("no log level: %v", name))
}
