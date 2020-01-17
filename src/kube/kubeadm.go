package kube

import (
	"context"
	"deployfromgo/src/config"
	"deployfromgo/src/logger"
	"errors"
	"fmt"
	"github.com/Lvzhenqian/sshtool"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"golang.org/x/crypto/ssh"
	"gopkg.in/yaml.v2"
	"math/rand"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"
)


var (
	TmpDir = "/tmp/kubeBuild"
	InitShellPath = "./k8s/init/env.sh"
	KubeadmShellPath = "./k8s/init/kubeadm.sh"
	ProxyImage = "ggangelo/apiproxy:v1.0"
	Ssh sshtool.SshClient
	Cfg *config.TomlConfig
	Masters []string
)

type EmptyType struct {}

func StringSet(s []string) (r []string) {
	var empyt EmptyType
	tmp := make(map[string]EmptyType)
	for _,v := range s{
		tmp[v] = empyt
	}
	for k := range tmp{
		r = append(r,k)
	}
	return
}

func init() {
	Ssh = new(sshtool.SSHTerminal)
	Cfg = config.Configmaps
	os.MkdirAll(TmpDir,0755)
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

func makeInit(ip string) error {
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

func makeKubeadmConfig() string {
	var sans []string
	l := strings.Split(Cfg.Kubeconf.LoadBalancer,":")
	copy(sans,Masters)
	sans = append(sans,l[0])
	init,proxy,kubelet := DefaultConfig()
	init.ApiServer = Apiserver{
		ExtraArgs: map[string]string{
			"authorization-mode":"Node,RBAC",
			"service-node-port-range": Cfg.Kubeconf.NodePortRang,
		},
		CertSANs:               StringSet(sans),
		TimeoutForControlPlane: "4m0s",
	}
	file := path.Join(TmpDir,"k8s.yaml")
	yml,CreateERR := os.Create(file)
	defer yml.Close()
	if CreateERR != nil{
		logger.Errorf("创建临时文件[%s] 失败！！%s",file,CreateERR)
		panic(CreateERR)
	}
	encode := yaml.NewEncoder(yml)
	defer encode.Close()
	encode.Encode(&init)
	encode.Encode(&proxy)
	encode.Encode(&kubelet)
	return file
}

func initCluster(ip string) (token string,certhash string) {
	configpath := makeKubeadmConfig()
	cli := newSSh(ip)
	defer cli.Close()
	Ssh.Push(configpath,"/tmp",cli)
	RunShell("kubeadm init --config /tmp/k8s.yaml",cli)
	RunCmd("mkdir -p /root/.kube",cli)
	RunCmd("/bin/cp -rf /etc/kubernetes/admin.conf /root/.kube/config",cli)
	ret := RunCmd("kubeadm token create --ttl 0 --print-join-command",cli)

	r := strings.Split(ret," ")
	index := func(v []string, s string) (int,error) {
		for i , v := range v{
			if v == s{
				return i,nil
			}
		}
		return -1,errors.New("找不到对应的值")
	}
	TokenIndex,_ := index(r,"--token")
	CerthashIndex,_ := index(r,"--discovery-token-ca-cert-hash")
	token = r[TokenIndex+1]
	certhash = r[CerthashIndex+1]
	return
}

func kubeconfig() {
	for _,ip := range Masters{
		go func(ip string) {
			cli := newSSh(ip)
			defer cli.Close()
			RunCmd("mkdir -p /root/.kube",cli)
			RunCmd("/bin/cp -rf /etc/kubernetes/admin.conf /root/.kube/config",cli)
			RunCmd("chown root.root $HOME/.kube/config",cli)
		}(ip)
	}
}

func forwardSameFile(paths string, src, dst *ssh.Client) error {
	return Ssh.Forward(paths,paths,src,dst)
}

func sendCrts(master string, other []string) error {
	const (
		ca = "/etc/kubernetes/pki/ca.crt"
		cakey = "/etc/kubernetes/pki/ca.key"
		sakey = "/etc/kubernetes/pki/sa.key"
		sa = "/etc/kubernetes/pki/sa.pub"
		front = "/etc/kubernetes/pki/front-proxy-ca.crt"
		frontkey = "/etc/kubernetes/pki/front-proxy-ca.key"
		etcd = "/etc/kubernetes/pki/etcd/ca.crt"
		etcdkey = "/etc/kubernetes/pki/etcd/ca.key"
	)
	srcCli := newSSh(master)
	defer srcCli.Close()
	for _,addr := range other{
		go func(src *ssh.Client,ip string) {
			dst := newSSh(ip)
			defer dst.Close()
			RunCmd("mkdir -p /etc/kubernetes/pki/etcd",dst)
			forwardSameFile(ca,src,dst)
			forwardSameFile(cakey,src,dst)
			forwardSameFile(sa,src,dst)
			forwardSameFile(sakey,src,dst)
			forwardSameFile(front,src,dst)
			forwardSameFile(frontkey,src,dst)
			forwardSameFile(etcd,src,dst)
			forwardSameFile(etcdkey,src,dst)
		}(srcCli,addr)
	}
	return nil
}

func taintMaster(ip, name string) error {
	cli := newSSh(ip)
	defer cli.Close()
	cmd := fmt.Sprintf("kubectl taint nodes %s node-role.kubernetes.io/master:NoSchedule-",name)
	return RunShell(cmd,cli)
}

func joinMaster(ip string) error {
	cli := newSSh(ip)
	defer cli.Close()
	cmd := fmt.Sprintf(
		"kubeadm join %s --token %s --discovery-token-ca-cert-hash %s --experimental-control-plane",
		ip,Cfg.Kubeconf.Token,Cfg.Kubeconf.CertHash)
	return RunShell(cmd,cli)
}

func joinNode(ip string) error {
	cli := newSSh(ip)
	defer cli.Close()
	cmd := fmt.Sprintf(
		"kubeadm join %s --token %s --discovery-token-ca-cert-hash %s",
		ip,Cfg.Kubeconf.Token,Cfg.Kubeconf.CertHash)
	return RunShell(cmd,cli)
}

func MakeInitServer()  {
	logger.Info("初始化服务器，结束后会把相应的服务器重启!!")
	var (
		wg sync.WaitGroup
		dolist sort.StringSlice
	)
	for _ ,v := range Cfg.Node{
		wg.Add(1)
		dolist = append(dolist,v)
		go func(w *sync.WaitGroup) {
			makeInit(v)
			w.Done()
		}(&wg)
	}
	wg.Wait()
	sort.Sort(sort.Reverse(dolist))
	for _,ip := range dolist{
		go restartServer(ip)
	}
}

