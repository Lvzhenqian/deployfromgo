package logger

import (
	"github.com/op/go-logging"
	"os"
)

const logfile string = "./install.log"

var (
	FileLog = logging.MustGetLogger("FileLog")
	//CLog = logging.MustGetLogger("Clog")
	format = logging.MustStringFormatter(
		"%{color}%{time:15:04:05.000} %{callpath} ▶ %{level:.4s} %{shortfile}%{color:reset} %{message}")
)


func init() {
	//ClogBackend := logging.NewLogBackend(os.Stdout,"",0)
	FileLogBackend := logging.NewLogBackend(FileWriter(logfile),"",0)
	Level := logging.AddModuleLevel(FileLogBackend)
	Level.SetLevel(logging.DEBUG,"")
	Format := logging.NewBackendFormatter(Level,format)
	logging.SetBackend(Format)
}

func FileWriter(filename string) *os.File {
	files,err := os.Create(filename)
	defer files.Close()
	if err != nil{
		panic(err)
	}
	return files
}
