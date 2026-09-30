package client

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/crypto/ssh"
)

func generateTestSSHKeys(t *testing.T) (ssh.Signer, string) {
	t.Helper()
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("failed to generate rsa key: %v", err)
	}

	privDER := x509.MarshalPKCS1PrivateKey(privateKey)
	privBlock := &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: privDER,
	}
	privPEM := pem.EncodeToMemory(privBlock)

	signer, err := ssh.ParsePrivateKey(privPEM)
	if err != nil {
		t.Fatalf("failed to parse private key: %v", err)
	}

	pubKey := signer.PublicKey()
	pubKeyAuthorized := string(ssh.MarshalAuthorizedKey(pubKey))

	return signer, pubKeyAuthorized
}

func startTestSSHServer(t *testing.T, signer ssh.Signer) (string, func()) {
	t.Helper()
	serverConfig := &ssh.ServerConfig{
		NoClientAuth: true,
	}
	serverConfig.AddHostKey(signer)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to listen on localhost: %v", err)
	}

	done := make(chan struct{})
	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				select {
				case <-done:
					return
				default:
					return
				}
			}
			go func(c net.Conn) {
				sshConn, chans, reqs, err := ssh.NewServerConn(c, serverConfig)
				if err != nil {
					return
				}
				defer sshConn.Close()
				go ssh.DiscardRequests(reqs)
				for newChan := range chans {
					ch, reqs, err := newChan.Accept()
					if err != nil {
						continue
					}
					go func(in <-chan *ssh.Request) {
						for req := range in {
							if req.Type == "exec" {
								_ = req.Reply(true, nil)
								_, _ = ch.Write([]byte("ok\n"))
								ch.SendRequest("exit-status", false, []byte{0, 0, 0, 0})
								_ = ch.Close()
								return
							}
							_ = req.Reply(false, nil)
						}
					}(reqs)
				}
			}(conn)
		}
	}()

	addr := listener.Addr().String()
	cleanup := func() {
		close(done)
		_ = listener.Close()
	}

	return addr, cleanup
}

func TestSSHClient_HostKeyVerification(t *testing.T) {
	signer, pubKeyAuth := generateTestSSHKeys(t)
	addr, cleanup := startTestSSHServer(t, signer)
	defer cleanup()

	host, portStr, _ := net.SplitHostPort(addr)
	var port int
	fmt.Sscanf(portStr, "%d", &port)

	t.Run("fails without known hosts and without insecure flag", func(t *testing.T) {
		tempHome := t.TempDir()
		t.Setenv("HOME", tempHome)

		cli := NewSSHClient(SSHOptions{
			Host:    host,
			Port:    port,
			Timeout: 2 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := cli.Connect(ctx)
		if err == nil {
			t.Fatal("expected connection to fail due to missing known_hosts file")
		}
		if !strings.Contains(err.Error(), "no known_hosts file found") {
			t.Fatalf("unexpected error message: %v", err)
		}
	})

	t.Run("succeeds with InsecureIgnoreHostKey", func(t *testing.T) {
		cli := NewSSHClient(SSHOptions{
			Host:                  host,
			Port:                  port,
			InsecureIgnoreHostKey: true,
			Timeout:               2 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := cli.Connect(ctx)
		if err != nil {
			t.Fatalf("expected successful connection with InsecureIgnoreHostKey: %v", err)
		}
		defer cli.Close()

		out, err := cli.RunCommand(ctx, "echo test")
		if err != nil {
			t.Fatalf("failed to run command: %v", err)
		}
		if strings.TrimSpace(out) != "ok" {
			t.Fatalf("unexpected output: %q", out)
		}
	})

	t.Run("succeeds with valid KnownHostsFile", func(t *testing.T) {
		tempDir := t.TempDir()
		knownHostsFile := filepath.Join(tempDir, "known_hosts")
		knownHostLine := fmt.Sprintf("[%s]:%d %s", host, port, strings.TrimSpace(pubKeyAuth))
		if err := os.WriteFile(knownHostsFile, []byte(knownHostLine+"\n"), 0600); err != nil {
			t.Fatalf("failed to write known_hosts: %v", err)
		}

		cli := NewSSHClient(SSHOptions{
			Host:           host,
			Port:           port,
			KnownHostsFile: knownHostsFile,
			Timeout:        2 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := cli.Connect(ctx)
		if err != nil {
			t.Fatalf("expected successful connection with valid known_hosts: %v", err)
		}
		defer cli.Close()
	})

	t.Run("fails when host key does not match known_hosts", func(t *testing.T) {
		tempDir := t.TempDir()
		knownHostsFile := filepath.Join(tempDir, "known_hosts")
		_, otherPubKeyAuth := generateTestSSHKeys(t)
		knownHostLine := fmt.Sprintf("[%s]:%d %s", host, port, strings.TrimSpace(otherPubKeyAuth))
		if err := os.WriteFile(knownHostsFile, []byte(knownHostLine+"\n"), 0600); err != nil {
			t.Fatalf("failed to write known_hosts: %v", err)
		}

		cli := NewSSHClient(SSHOptions{
			Host:           host,
			Port:           port,
			KnownHostsFile: knownHostsFile,
			Timeout:        2 * time.Second,
		})

		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()

		err := cli.Connect(ctx)
		if err == nil {
			t.Fatal("expected connection to fail due to host key mismatch")
		}
	})
}
