package logger

import "testing"


func TestDefaultInfo(t *testing.T) {
	Error("err!!!")
	Info("Info")
	Debug("Debug")
	Critical("critical")
	Notice("Notice")
	Warning("Warning!")
}