// Package ssh wraps core ssh package
// this package is still in development
package ssh

import (
	"fmt"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"io/ioutil"
	"net"
	"os/user"
	"strings"
	"time"

	"github.com/mehrdadrad/mylg/cli"
)

// SSH represents SSH properties
type SSH struct {
	Username  string
	Password  string
	PublicKey string
	Host      string
	Keepalive string
	Timeout   time.Duration
	Config    *ssh.ClientConfig
	Client    *ssh.Client
	Session   *ssh.Session
}

// Conn represents connection and timeouts
type Conn struct {
	net.Conn
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Read set read timeout
func (c *Conn) Read(b []byte) (int, error) {
	err := c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Read(b)
}

// Write set write timeout
func (c *Conn) Write(b []byte) (int, error) {
	err := c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
	if err != nil {
		return 0, err
	}
	return c.Conn.Write(b)
}

// ClientConfig adds auth methods
func (s *SSH) ClientConfig() error {
	var auths []ssh.AuthMethod

	if s.Password != "" {
		auths = append(auths, ssh.Password(s.Password))
	}

	if s.PublicKey != "" {
		pem, err := ioutil.ReadFile(s.PublicKey)
		if err != nil {
			return err
		}
		signer, err := ssh.ParsePrivateKey(pem)
		if err != nil {
			return err
		}
		auths = append(auths, ssh.PublicKeys(signer))
	}

	s.Config = &ssh.ClientConfig{
		User: s.Username,
		Auth: auths,
	}
	return nil
}

// NewClient make new SSH client
func (s *SSH) NewClient() error {

	// connection
	conn, err := net.DialTimeout("tcp", s.Host, s.Timeout)
	if err != nil {
		return err
	}
	c, ch, req, err := ssh.NewClientConn(
		&Conn{conn, s.Timeout, s.Timeout},
		s.Host,
		s.Config,
	)
	if err != nil {
		return err
	}

	// init client
	s.Client = ssh.NewClient(c, ch, req)

	// keep alive
	go func() {
		keepalive, err := time.ParseDuration(s.Keepalive)
		if err != nil {
			println("keep alive format incorrect")
			return
		}
		t := time.NewTicker(keepalive)
		defer t.Stop()
		for {
			<-t.C
			_, _, err := s.Client.Conn.SendRequest("keepalive@mylg.io", true, nil)
			if err != nil {
				return
			}
		}
	}()

	return nil
}

// NewPty creates new session, terminal and request pty
func (s *SSH) NewPty() error {
	session, err := s.Client.NewSession()
	if err != nil {
		return err
	}
	// set terminal
	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}
	// request pty
	if err := session.RequestPty("xterm", 80, 40, modes); err != nil {
		return err
	}

	s.Session = session
	return nil
}

// NewSSH creates new ssh object based on the cli
func NewSSH(args string, cfg cli.Config) (*SSH, error) {
	var (
		username string
		port     = "22"
	)

	target, flag := cli.Flag(args)
	if _, ok := flag["help"]; ok || len(target) < 3 {
		help(cfg)
		return nil, nil
	}

	// port
	if p, ok := flag["-p"]; ok {
		port = p.(string)
	} else if h, p, err := net.SplitHostPort(target); err == nil {
		port = p
		target = h
	}

	// user and host
	if strings.Contains(target, "@") {
		t := strings.Split(target, "@")
		username = t[0]
		target = t[1]
	} else {
		usr, err := user.Current()
		if err != nil {

		}
		username = usr.Username
	}

	host := net.JoinHostPort(target, port)

	// password
	fmt.Printf("please enter %s@%s's password: ", username, target)
	password, err := terminal.ReadPassword(0)
	if err != nil {

	}

	return &SSH{
		Username:  username,
		Password:  string(password),
		Host:      host,
		Keepalive: "2s",
		Timeout:   5 * time.Second,
	}, nil

}

// help
func help(cfg cli.Config) {
	// TODO
}
