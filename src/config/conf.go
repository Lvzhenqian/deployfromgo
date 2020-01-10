package config

import (
	"github.com/BurntSushi/toml"
	"os"
	"path/filepath"
	"sync"
)

const ConfigPath string = "../../conf/test.toml"

var Configmaps  *TomlConfig

type TomlConfig struct {
	Kubeconf Kubeconf
	Node 	map[string]string
	Ssh 	SshConf
}

type SshConf struct {
	Port 		int
	Username	string
	Password	string
	PkeyPath	string
}

type Kubeconf struct {
	Version			string
	ServiceCidr		string
	PodCidr			string
	DataDir 		string
	EtcdDir			string
	NodePortRang	string
	DockerVersion	string
	ProxyMode		string
	NetworkAddons 	string
	DashboardToken	string
	MTU				string
	Token			string
	CertHash		string
	LoadBalancer	string
}

func (conf *TomlConfig) Write(config string) error {
	file,_ := filepath.Abs(config)
	f,err := os.Create(file)
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

func NewConfig() *TomlConfig {
	return &TomlConfig{
		Kubeconf: Kubeconf{
			Version:"1.17.0",
			ServiceCidr: "10.254.0.0/16",
			PodCidr:"10.50.0.0/16",
			DataDir: "/data/kubernetes",
			EtcdDir:"/data/etcd",
			NodePortRang: "30000-50000",
			DockerVersion: "docker-ce-19.03.5-3.el7.x86_64",
			ProxyMode: "ipvs",
			NetworkAddons: "canal",
			MTU:"1440",
			LoadBalancer: "127.0.0.1:8443",
		},
		Node:     nil,
		Ssh:      SshConf{
				Port: 22,
		},
	}
}

func init() {
	Configmaps = NewConfig()
	err := Configmaps.Read(ConfigPath)
	if err != nil {
		panic(err)
	}
}