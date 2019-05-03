package config

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"sync"
)

var (
	Perfix string = "k8s-"
	LoadBalancer string = "127.0.0.1:5443"
)

type Master struct {
	IPS []string
}

type TomlConfig struct {
	Kubeconf Kubeconf
	Master 	 Master
	Node 	map[string]string
	Ssh 	SshConf
}

type SshConf struct {
	Port 		int
	Username	string
	Password	string
}

type Kubeconf struct {
	Version			string
	ServiceCidr		string
	PodCidr			string
	DataDir 		string
	NodePortRang	string
	DockerVersion	string
	ProxyMode		string
	NetworkAddons 	string
	MTU				string
	GrafanaPasswd 	string
	Token			string
	CertHash		string
	LoadBalancer	string
}

func (conf *TomlConfig) Write(config string) error {
	file,_ := filepath.Abs(config)
	f,err := os.Open(file)
	defer f.Close()
	if err != nil {
		return err
	}
	encode := toml.NewEncoder(f)
	if err := encode.Encode(&conf);err != nil{
		return err
	}
	return nil
}

func (conf *TomlConfig) Read(config string) error  {
	var once sync.Once
	once.Do(func()  {
		file,_ := filepath.Abs(config)
		if _, err := toml.DecodeFile(file,&conf) ; err !=nil {
			panic(err)
		}
	})
	return nil
}