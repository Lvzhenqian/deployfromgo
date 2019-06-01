package config

import (
	"reflect"
	"testing"
)

func assertType(t *testing.T,a interface{},b string) {
	ta := reflect.TypeOf(a)
	if ta.Name() != b {
		t.Errorf("Not Equal %v must:%v but %v \n",a,b,ta.Name())
	}
}


func TestTomlConfig_Read(t *testing.T) {
	assertType(t,Configmaps.Ssh.Port,"int")
}

func TestTomlConfig_Write(t *testing.T) {
	Configmaps.Ssh.Port = 3879
	err := Configmaps.Write("./1.toml")
	if err !=nil{
		t.Error(err)
	}
}