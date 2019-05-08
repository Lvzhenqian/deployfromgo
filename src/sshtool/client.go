package sshtool

import (
	"bufio"
	"deployfromgo/src/config"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"gopkg.in/cheggaaa/pb.v1"
	"io"
	"log"
	"net"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type SSH interface {
	ExecCommand(ip string,cmd string) error
	PushFile(ip string ,Src string,Dst string) error
	GetFile(ip string, Src string,Dst string) error
	PushDir(ip string, Src string,Dst string) error
	GetDir(ip string, Src string,Dst string)
}

var (
	Configs 	*config.TomlConfig
	ClientCf	*ssh.ClientConfig
	)

type SshClient struct{
	CmdOutPut string
}

func init() {
	Configs = new(config.TomlConfig)
	err := Configs.Read("../conf/test.toml")
	if err != nil {
		panic(err)
	}
	auth := make([]ssh.AuthMethod,0)
	auth = append(auth,ssh.Password(Configs.Ssh.Password))
	ClientCf = &ssh.ClientConfig{
		User: Configs.Ssh.Username,
		Auth: auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
}

func (c *SshClient) ExecCommand(ip string, cmd string) error {
	var Out strings.Builder
	//var Out bytes.Buffer
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,Configs.Ssh.Port), ClientCf)
	if clienterr != nil {
		log.Fatalf("Client Error: %v\n",clienterr)
		return clienterr
	}
	defer client.Close()
	session ,SessionErr := client.NewSession()
	defer session.Close()
	if SessionErr != nil{
		log.Fatalf("Session Error: %v\n",SessionErr)
		return SessionErr
	}
	reader, ReaderErr := session.StdoutPipe()
	if ReaderErr != nil{
		log.Fatalf("reader Error: %v\n",ReaderErr)
		return ReaderErr
	}
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			//n,e:=Out.Write(scanner.Bytes())
			//fmt.Println(n,e)
			fmt.Fprintf(&Out,"%s\n",scanner.Text())
		}
	}()

	if err:=session.Run(cmd); err != nil {
		return err
	}
	c.CmdOutPut = Out.String()
	//fmt.Println(c.CmdOutPut)
	return nil
}

func (c *SshClient) PushFile(ip string, Src string, Dst string) error {
	//create connect
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,Configs.Ssh.Port), ClientCf)
	if clienterr != nil {
		return clienterr
	}
	defer client.Close()
	// new client
	sftpClient, err := sftp.NewClient(client)
	defer sftpClient.Close()
	// open file
	srcFile, err := os.Open(Src)
	defer srcFile.Close()
	if err != nil {
		return err
	}
	dstFile, err := sftpClient.Create(Dst)
	defer dstFile.Close()
	//bar
	SrcStat,err := srcFile.Stat()
	if err != nil {
		return err
	}
	bar := pb.New64(SrcStat.Size()).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowTimeLeft=true
	bar.ShowPercent=true
	bar.Prefix(path.Base(Src))
	bar.Start()
	r:=bar.NewProxyReader(srcFile)
	defer bar.Finish()
	if _, err :=io.Copy(dstFile,r);err != nil {
		return err
	}

	return nil
}

func (c *SshClient) GetFile(ip string, Src string, Dst string) error {
	// create SshClient
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,Configs.Ssh.Port), ClientCf)
	if clienterr != nil {
		return clienterr
	}
	defer client.Close()
	// new SftpClient
	sftpClient, err := sftp.NewClient(client)
	defer sftpClient.Close()
	if err != nil {
		return err
	}
	// open SrcFile
	srcFile, err := sftpClient.Open(Src)
	defer srcFile.Close()
	if err != nil {
		return err
	}
	//bar
	SrcStat,err :=srcFile.Stat()
	if err != nil {
		return err
	}
	bar := pb.New64(SrcStat.Size()).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowTimeLeft=true
	bar.ShowPercent=true
	bar.Prefix(path.Base(Src))
	bar.Start()
	// open DstFile
	dstFile, err := os.Create(Dst)
	defer dstFile.Close()
	w := io.MultiWriter(bar,dstFile)
	defer bar.Finish()
	if _,err := srcFile.WriteTo(w);err != nil {
		return err
	}

	return nil
}

func (c *SshClient) PushDir(ip string, Src string, Dst string) error {
	// create SshClient
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,Configs.Ssh.Port), ClientCf)
	if clienterr != nil {
		return clienterr
	}
	defer client.Close()
	// new SftpClient
	sftpClient, err := sftp.NewClient(client)
	defer sftpClient.Close()
	if err != nil {
		return err
	}

	root,dir := path.Split(Src)
	if err:=os.Chdir(root);err != nil{
		return err
	}
	size := TotalSize(Src)
	bar := pb.New64(size).SetUnits(pb.U_BYTES)
	bar.ShowSpeed = true
	bar.ShowTimeLeft=true
	bar.ShowPercent=true
	bar.Prefix(path.Base(Src))
	bar.Start()
	defer bar.Finish()
	var wg sync.WaitGroup
	WalkErr := filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		switch {
		case info.IsDir():
			sftpClient.Mkdir(p)
		default:
			dstfile := path.Join(Dst,p)
			wg.Add(1)
			go func(wgroup *sync.WaitGroup,b *pb.ProgressBar, Srcfile string,Dstfile string) {
				defer wgroup.Done()
				s,_ := os.Open(Srcfile)
				defer s.Close()
				d,_ := sftpClient.Create(Dstfile)
				defer d.Close()
				i,_ :=io.Copy(d,s)
				b.Add64(i)
			}(&wg,bar,p,dstfile)
		}
		wg.Wait()
		return err
	})

	if WalkErr !=nil{
		return err
	}
	return nil
}

func (c *SshClient) GetDir(ip string, Src string, Dst string) {
	panic("implement me")
}


func TotalSize(paths string) int64 {
	var Ret int64
	stat,_ := os.Stat(paths)
	switch  {
	case stat.IsDir():
		filepath.Walk(paths, func(p string, info os.FileInfo, err error) error {
			if info.IsDir(){
				return nil
			} else {
				s,_ := os.Stat(p)
				Ret = Ret +s.Size()
				return nil
			}
		})
		return Ret
	default:
		return stat.Size()
	}
}
