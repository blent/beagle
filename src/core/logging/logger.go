package logging

import (
	"fmt"
	"io"
	"log"
	"os"
	"strings"
)

var DefaultOutput = os.Stdout

type (
	Logger struct {
		info  *log.Logger
		warn  *log.Logger
		error *log.Logger
		fatal *log.Logger
	}
)

func NewDefaultLogger(name string) *Logger {
	// log.SetFlags(log.Lshortfile)
	return NewLogger(name, DefaultOutput)
}

func NewLogger(name string, out io.Writer) *Logger {
	return &Logger{
		info:  log.New(out, fmt.Sprintf("[%s] [INFO] ", strings.ToUpper(name)), log.Ldate|log.Ltime),
		warn:  log.New(out, fmt.Sprintf("[%s] [WARN] ", strings.ToUpper(name)), log.Ldate|log.Ltime),
		error: log.New(out, fmt.Sprintf("[%s] [ERROR] ", strings.ToUpper(name)), log.Ldate|log.Ltime),
		fatal: log.New(out, fmt.Sprintf("[%s] [FATAL] ", strings.ToUpper(name)), log.Ldate|log.Ltime),
	}
}

func (logger *Logger) Print(v ...interface{}) {
	logger.info.Println(v...)
}

func (logger *Logger) Info(message string) {
	logger.info.Println(message)
}

func (logger *Logger) Infof(format string, v ...interface{}) {
	logger.info.Printf(format, v...)
}

func (logger *Logger) Warn(message string) {
	logger.warn.Println(message)
}

func (logger *Logger) Warnf(format string, v ...interface{}) {
	logger.warn.Printf(format, v...)
}

func (logger *Logger) Error(message string) {
	logger.error.Println(message)
}

func (logger *Logger) Errorf(format string, v ...interface{}) {
	logger.error.Printf(format, v...)
}

func (logger *Logger) Fatal(message string) {
	logger.fatal.Println(message)
}

func (logger *Logger) Fatalf(format string, v ...interface{}) {
	logger.fatal.Printf(format, v...)
}
