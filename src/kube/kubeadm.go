package kube

import (
	"context"
	"deployfromgo/src/config"
	"deployfromgo/src/logger"
	"fmt"
	"github.com/Lvzhenqian/sshtool"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/ssh"
	"math/rand"
	"strings"
	"time"
)


var (
	InitShellPath = "./k8s/init/env.sh"
	KubeadmShellPath = "./k8s/init/kubeadm.sh"
	ProxyImage = "ggangelo/apiproxy:v1.0"
	Ssh sshtool.SshClient
	Cfg *config.TomlConfig
	Masters []string

)

func init() {
	Ssh = new(sshtool.SSHTerminal)
	Cfg = config.Configmaps
}

func newSSh(ip string) *ssh.Client {
	client, NewERR := sshtool.NewClient(ip,Cfg.Ssh.Port,Cfg.Ssh.Username,Cfg.Ssh.Password,Cfg.Ssh.PkeyPath)
	if NewERR != nil{
		logger.Errorf("%s: 创建sshclient失败！！",ip)
		panic(NewERR)
	}
	return client
}

func RunShell(cmd string, c *ssh.Client) error {
	write := logger.NewReadWriteDebugPipe()
	return Ssh.Run(cmd,write,c)
}

func FindKeyFromValue(value string,m map[string]string) string {
	tmp := make(map[string]string)
	for k,v := range m{
		tmp[v] = k
	}
	return tmp[value]
}

func RunCmd(cmd string, c *ssh.Client) string {
	s := new(strings.Builder)
	e := Ssh.Run(cmd,s,c)
	if e != nil {
		logger.Error(e)
	}
	return s.String()
}

func Init(ip string) error {
	client := newSSh(ip)
	defer client.Close()
	if PushErr := Ssh.Push(InitShellPath,"/tmp",client); PushErr != nil{
		logger.Errorf("发送[%s]文件到/tmp目录失败",InitShellPath)
		panic(PushErr)
	}
	return RunShell("/bin/bash /tmp/env.sh",client)
}

func restartServer(ip string) error {
	logger.Infof("[%s] 服务器将要重启!!",ip)
	c := newSSh(ip)
	defer c.Close()
	write := logger.NewReadWriteDebugPipe()
	return Ssh.Run("reboot",write,c)
}

func kubeadm(ip string) error {
	c := newSSh(ip)
	defer c.Close()
	if err :=Ssh.Push(KubeadmShellPath,"/tmp",c);err!=nil{
		logger.Errorf("发送[%s]文件到/tmp目录失败",KubeadmShellPath)
		panic(err)
	}
	return RunShell("/bin/bash /tmp/kubeadm.sh",c)
}

func setHostName(ip string) error {
	var hosts []string
	hostname := FindKeyFromValue(ip,Cfg.Node)
	setname := fmt.Sprintf("hostnamectl set-hostname %s",hostname)
	for name,node := range Cfg.Node{
		hosts = append(hosts,fmt.Sprintf("%s %s",node,name))
	}
	hostlist := strings.Join(hosts,"\n")
	headers := `127.0.0.1   localhost localhost.localdomain localhost4 localhost4.localdomain4
::1         localhost localhost.localdomain localhost6 localhost6.localdomain6`
	localhosts := fmt.Sprintf("echo \"%s\n%s\" > /etc/hosts",headers,hostlist)
	backup := "/bin/cp -rf /etc/hosts /etc/hosts_old"
	cli := newSSh(ip)
	defer cli.Close()
	for _,value := range []string{backup,setname,localhosts,}{
		if err := RunShell(value,cli);err != nil{
			logger.Error(err)
			return err
		}
	}
	return nil
}

func makeProxyFromDockerSdk(ip string) error {
	// 定义随机端口
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	LocalAddress := fmt.Sprintf("%s:%d","127.0.0.1",6000+r.Intn(100))
	localTunnel := sshtool.TunnelSetting{
		Network: "tcp",
		Address: LocalAddress,
	}
	logger.Debugf("%s 打开的端口：%v",ip,localTunnel)
	dockerTunnel := sshtool.TunnelSetting{
		Network: "unix",
		Address: "/var/run/docker.sock",
	}
	cli := newSSh(ip)
	defer cli.Close()
	// 创建docker的tunnel，然后使用tunnel创建对应的客户端
	go Ssh.TunnelStart(localTunnel,dockerTunnel,cli)
	dockerLocalPort := fmt.Sprintf("%s://%s",localTunnel.Network,localTunnel.Address)
	dockerCli,DockerErr := client.NewClient(dockerLocalPort,client.DefaultVersion,nil,nil)
	if DockerErr != nil{
		logger.Errorf("创建docker客户端失败")
		panic(DockerErr)
	}
	defer dockerCli.Close()
	// 拉取映像
	ctx := context.Background()
	reader,Pullerr := dockerCli.ImagePull(ctx,ProxyImage,types.ImagePullOptions{})
	if Pullerr != nil{
		logger.Errorf("%s pull % Error!!",ip,ProxyImage)
		return Pullerr
	}
	logger.DebugFromReader(fmt.Sprintf("%s: ",ProxyImage),reader)

	// 把所有master地址api接口放到变量a里面，生成对应的代理列表
	var a []string
	for _,x := range Masters{
		a = append(a,x+":6443")
	}
	m := strings.Join(a,",")
    // PortMap
	listen := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: "8443",
	}
	admin := nat.PortBinding{
		HostIP:   "0.0.0.0",
		HostPort: "9000",
	}
	// 创建代理容器
	con,CrErr:=dockerCli.ContainerCreate(ctx,&container.Config{
		Env:            []string{"LISTEN=:8443",fmt.Sprintf("BACKEND=%s",m),},
		Image:           ProxyImage,
		Labels: map[string]string{"app":"proxy",},

	},&container.HostConfig{
		PortBindings:    nat.PortMap{
			"8443":[]nat.PortBinding{listen,},
			"9000":[]nat.PortBinding{admin,},
		},
		RestartPolicy:   container.RestartPolicy{Name: "always",},
	},nil,"apiserver-proxy")
	if CrErr != nil{
		logger.Errorf("创建apiserver-proxy失败： %s",CrErr)
		return CrErr
	}
	logger.Debugf("proxy：%s",con.ID)
	return nil
}

//func makeKubeadmConfig() string {
//
//}