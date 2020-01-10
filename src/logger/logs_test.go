package logger

import "testing"


func TestDefaultInfo(t *testing.T) {
	Error("err!!!")
	Info("Info")
	Debug("Debug")
	Critical("critical")
	Notice("Notice")
	Warning("Warning!")

	Errorf("err %s","hello")
	Infof("err %s","hello")
	Debugf("err %s","hello")
	Criticalf("err %s","hello")
	Noticef("err %s","hello")
	Warningf("err %s","hello")
}