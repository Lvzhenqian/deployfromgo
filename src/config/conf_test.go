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
