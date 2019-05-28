package logger

import (
	"github.com/op/go-logging"
	"io"
	"os"
)

type Log struct {
	file  			*logging.Logger
	FileLevel 		int
	console 		*logging.Logger
	ConsoleLevel 	int
}

const LOGFILE string = "./install.log"

var logs *Log

func SetLoger(w io.Writer,logName string,level int,fmts string) *logging.Logger {
	backend := logging.NewLogBackend(w,"",0)
	format := logging.NewBackendFormatter(backend,logging.MustStringFormatter(fmts))
	backendlevel := logging.AddModuleLevel(format)
	backendlevel.SetLevel(logging.Level(level),logName)
	l := logging.MustGetLogger(logName)
	l.SetBackend(backendlevel)
	return l
}

func InitLogger() *Log {
	e := new(Log)
	e.ConsoleLevel = 3
	e.FileLevel = 5
	// log file
	FileWrite,OpenErr := os.OpenFile(LOGFILE,os.O_RDWR|os.O_CREATE|os.O_APPEND,0666)
	if OpenErr != nil {
		panic(OpenErr)
	}
	fmtfile := "%{color}%{time:15:04:05.000} %{callpath} ▶  %{level:.4s} %{shortfile}%{color:reset} %{message}"
	e.file = SetLoger(FileWrite,"FileLoger",e.FileLevel,fmtfile)

	// Console
	cfmt := "%{color}%{message}%{color:reset}"
	e.console = SetLoger(os.Stdout,"Console",e.ConsoleLevel,cfmt)
	return e
}

func init() {
	logs = InitLogger()
}

func (l *Log) Info(args ...interface{}) {
	l.file.Info(args...)
}

func (l *Log) Cinfo(args ...interface{}) {
	l.console.Info(args...)
}

func Error(args ...interface{}) {
	logs.file.Error(args...)
	logs.console.Error(args...)
}

func Info(args ...interface{}) {
	logs.file.Info(args...)
	logs.console.Info(args...)
}
func Debug(args ...interface{}) {
	logs.file.Debug(args...)
	logs.console.Debug(args...)
}

func Warning(args ...interface{})  {
	logs.file.Warning(args...)
	logs.console.Warning(args...)
}

func Notice(args ...interface{}) {
	logs.file.Notice(args...)
	logs.console.Notice(args...)
}

func Critical(args ...interface{}) {
	logs.file.Critical(args...)
	logs.console.Critical(args...)
}

