package main

import (
	"fmt"
	"github.com/Lvzhenqian/sshtool"
)

func main() {
	//var c sshtool.SSH
	var d sshtool.SshClient

	err := d.GetDir("","/data/jupyter","d:/")
	if err != nil {
		fmt.Println(err)
	}
	//c = &d
	//err:= d.ExecCommand("192.168.3.127","w")
	//if perr != nil {
	//	log.Fatal(perr)
	//}
	//fmt.Println(d.CmdOutPut)
	//if err != nil {
	//	panic(err)
	//}
}