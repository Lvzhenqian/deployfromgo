package sshtool

import (
	"bufio"
	"deployfromgo/src/config"
	"fmt"
	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"net"
	"os"
	"strings"
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
	SshConfig config.SshConf
	ClientCf	*ssh.ClientConfig
	)

type SshClient struct{
	CmdOutPut string
}

func init() {
	auth := make([]ssh.AuthMethod,0)
	auth = append(auth,ssh.Password(SshConfig.Password))
	ClientCf = &ssh.ClientConfig{
		User: SshConfig.Username,
		Auth: auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
}

func (c *SshClient) ExecCommand(ip string, cmd string) error {
	var Out strings.Builder
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,SshConfig.Port), ClientCf)
	if clienterr != nil {
		return clienterr
	}
	defer client.Close()
	session ,SessionErr := client.NewSession()
	defer session.Close()
	if SessionErr != nil{
		return clienterr
	}
	reader, ReaderErr := session.StdoutPipe()
	if ReaderErr != nil{
		return clienterr
	}
	scanner := bufio.NewScanner(reader)
	go func() {
		for scanner.Scan() {
			fmt.Fprintf(&Out,"&s",scanner.Text())
		}
	}()
	c.CmdOutPut = Out.String()
	return session.Run(cmd)
}

func (c *SshClient) PushFile(ip string, Src string, Dst string) error {
	//create connect
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,SshConfig.Port), ClientCf)
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
	Scanner := bufio.NewScanner(srcFile)
	for Scanner.Scan() {
		_,err := dstFile.Write(Scanner.Bytes())
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *SshClient) GetFile(ip string, Src string, Dst string) error {
	// create SshClient
	client,clienterr := ssh.Dial("tcp", fmt.Sprintf("%s:%d",ip,SshConfig.Port), ClientCf)
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
	// open DstFile
	dstFile, err := os.Create(Dst)
	defer dstFile.Close()
	if _,err := srcFile.WriteTo(dstFile);err != nil {
		return err
	}
	return nil
}

func (c *SshClient) PushDir(ip string, Src string, Dst string) error {
	panic("implement me")
}

func (c *SshClient) GetDir(ip string, Src string, Dst string) {
	panic("implement me")
}

