package logger

import (
	"github.com/op/go-logging"
	"os"
)

const logfile string = "./install.log"

var (
	FileLog = logging.MustGetLogger("FileLog")
	Fileformat = logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{callpath} ▶ %{level:.4s} %{shortfile}%{color:reset} %{message}")
	Console = logging.MustGetLogger("ConsoleLog")
	ConsoleFormat = logging.MustStringFormatter("%{color}%{message}%{color:reset}")
	)


func init() {
	// log file
	FileWrite,OpenErr := os.OpenFile(logfile,os.O_RDWR|os.O_CREATE,777)
	if OpenErr != nil {
		panic(OpenErr)
	}
	FileLogBackend := logging.NewLogBackend(FileWrite,"",0)
	FileLevel := logging.AddModuleLevel(FileLogBackend)
	FileLevel.SetLevel(logging.DEBUG,"FileLog")
	FilesFormat := logging.NewBackendFormatter(FileLevel,Fileformat)
	// Console
	ConsoleBackend := logging.NewLogBackend(os.Stdout,"",0)
	ConsoleLevel := logging.AddModuleLevel(ConsoleBackend)
	ConsoleLevel.SetLevel(logging.INFO,"ConsoleLog")
	consolesFormat := logging.NewBackendFormatter(ConsoleBackend,ConsoleFormat)
	logging.SetBackend(FilesFormat,consolesFormat)
}

