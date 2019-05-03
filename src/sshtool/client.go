package sshtool

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"net"
	"time"
)

type Client struct {
	session 		*ssh.Session
}

func (c *Client) ConnectFromPasswd(username string, password string, host string, port int) (*ssh.Session, error) {
	auth := make([]ssh.AuthMethod,0)
	auth = append(auth, ssh.Password(password))
	ClientConfig := &ssh.ClientConfig{
		User: username,
		Auth: auth,
		Timeout: 30 * time.Second,
		HostKeyCallback: func(hostname string, remote net.Addr, key ssh.PublicKey) error {
			return nil
		},
	}
	address := fmt.Sprintf("%s:%d",host,port)
	if client, err := ssh.Dial("tcp",address,ClientConfig); err != nil {
		return nil,err
	} else {
		c.session, _ = client.NewSession()
		return c.session,err
	}
}