package logger

import (
	"bufio"
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

func NewReadWriteDebugPipe() *io.PipeWriter {
	r,w := io.Pipe()
	go func() {
		scan := bufio.NewScanner(r)
		for scan.Scan(){
			v := scan.Text()
			logs.file.Debug(v)
			logs.console.Debug(v)
		}
	}()
	return w
}

func DebugFromReader(title string,r io.Reader) {
	logs.file.Debug(title)
	scan := bufio.NewScanner(r)
	for scan.Scan(){
		logs.file.Debug(scan.Text())
	}
}

func Error(args ...interface{}) {
	logs.file.Error(args...)
	logs.console.Error(args...)
}

func Errorf(formats string, args ...interface{}) {
	logs.file.Errorf(formats,args...)
	logs.console.Errorf(formats,args...)
}

func Info(args ...interface{}) {
	logs.file.Info(args...)
	logs.console.Info(args...)
}

func Infof(formats string, args ...interface{}) {
	logs.file.Infof(formats,args...)
	logs.console.Infof(formats,args...)
}

func Debug(args ...interface{}) {
	logs.file.Debug(args...)
	logs.console.Debug(args...)
}

func Debugf(formats string,args ...interface{})  {
	logs.file.Debugf(formats,args...)
	logs.console.Debugf(formats,args...)
}

func Warning(args ...interface{})  {
	logs.file.Warning(args...)
	logs.console.Warning(args...)
}

func Warningf(formats string, args ...interface{}) {
	logs.file.Warningf(formats,args...)
	logs.console.Warningf(formats,args...)
}

func Notice(args ...interface{}) {
	logs.file.Notice(args...)
	logs.console.Notice(args...)
}

func Noticef(formats string, args ...interface{}) {
	logs.file.Noticef(formats,args...)
	logs.console.Noticef(formats,args...)
}

func Critical(args ...interface{}) {
	logs.file.Critical(args...)
	logs.console.Critical(args...)
}

func Criticalf(formats string, args ...interface{}) {
	logs.file.Criticalf(formats,args...)
	logs.console.Criticalf(formats,args...)
}

