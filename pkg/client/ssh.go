package client

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"os"
	"strings"
	"strconv"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHOptions struct {
	Host       string
	Port       int
	Username   string
	Password   string
	KeyPath    string
	KeyData    []byte
	Timeout    time.Duration
	Passphrase string
}

type SSHClient struct {
	opts   SSHOptions
	client *ssh.Client
}

func NewSSHClient(opts SSHOptions) *SSHClient {
	if opts.Port <= 0 {
		opts.Port = 22
	}
	if opts.Timeout <= 0 {
		opts.Timeout = 15 * time.Second
	}
	return &SSHClient{opts: opts}
}

func (s *SSHClient) Connect(ctx context.Context) error {
	var authMethods []ssh.AuthMethod

	if s.opts.Password != "" {
		authMethods = append(authMethods, ssh.Password(s.opts.Password))
	}

	var keyBytes []byte
	if len(s.opts.KeyData) > 0 {
		keyBytes = s.opts.KeyData
	} else if s.opts.KeyPath != "" {
		data, err := os.ReadFile(s.opts.KeyPath)
		if err == nil {
			keyBytes = data
		}
	}

	if len(keyBytes) > 0 {
		var signer ssh.Signer
		var err error
		if s.opts.Passphrase != "" {
			signer, err = ssh.ParsePrivateKeyWithPassphrase(keyBytes, []byte(s.opts.Passphrase))
		} else {
			signer, err = ssh.ParsePrivateKey(keyBytes)
		}
		if err == nil {
			authMethods = append(authMethods, ssh.PublicKeys(signer))
		}
	}

	config := &ssh.ClientConfig{
		User:            s.opts.Username,
		Auth:            authMethods,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         s.opts.Timeout,
	}

	addr := net.JoinHostPort(s.opts.Host, strconv.Itoa(s.opts.Port))

	var d net.Dialer
	conn, err := d.DialContext(ctx, "tcp", addr)
	if err != nil {
		return err
	}

	c, chans, reqs, err := ssh.NewClientConn(conn, addr, config)
	if err != nil {
		conn.Close()
		return err
	}

	s.client = ssh.NewClient(c, chans, reqs)
	return nil
}

func (s *SSHClient) Close() error {
	if s.client != nil {
		return s.client.Close()
	}
	return nil
}

func (s *SSHClient) RunCommand(ctx context.Context, cmd string) (string, error) {
	if s.client == nil {
		return "", fmt.Errorf("ssh client not connected")
	}

	sess, err := s.client.NewSession()
	if err != nil {
		return "", err
	}
	defer sess.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	sess.Stdout = &stdoutBuf
	sess.Stderr = &stderrBuf

	done := make(chan error, 1)
	go func() {
		done <- sess.Run(cmd)
	}()

	select {
	case <-ctx.Done():
		_ = sess.Signal(ssh.SIGKILL)
		return "", ctx.Err()
	case err := <-done:
		if err != nil {
			return strings.TrimSpace(stdoutBuf.String()), fmt.Errorf("%w: %s", err, stderrBuf.String())
		}
		return strings.TrimSpace(stdoutBuf.String()), nil
	}
}
