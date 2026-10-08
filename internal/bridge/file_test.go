package bridge

import (
	"bytes"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/pkg/sftp"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
)

func syntheticFileSession(t *testing.T) *Session {
	t.Helper()
	signer := testSigner(t)
	config := &ssh.ServerConfig{NoClientAuth: true}
	config.AddHostKey(signer)
	pair, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_STREAM, 0)
	if err != nil {
		t.Fatal(err)
	}
	connections := make([]net.Conn, 2)
	for i, fd := range pair {
		file := os.NewFile(uintptr(fd), "synthetic-sftp")
		connections[i], err = net.FileConn(file)
		file.Close()
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { connections[i].Close() })
		connections[i].SetDeadline(time.Now().Add(5 * time.Second))
	}
	go func() {
		server, channels, requests, err := ssh.NewServerConn(connections[1], config)
		if err != nil {
			return
		}
		defer server.Close()
		go ssh.DiscardRequests(requests)
		for incoming := range channels {
			if incoming.ChannelType() != "session" {
				incoming.Reject(ssh.UnknownChannelType, "session required")
				continue
			}
			channel, requests, err := incoming.Accept()
			if err != nil {
				return
			}
			go func() {
				defer channel.Close()
				for request := range requests {
					var payload struct{ Name string }
					if request.Type != "subsystem" || ssh.Unmarshal(request.Payload, &payload) != nil || payload.Name != "sftp" {
						request.Reply(false, nil)
						continue
					}
					request.Reply(true, nil)
					server, err := sftp.NewServer(channel)
					if err != nil {
						return
					}
					defer server.Close()
					server.Serve()
					return
				}
			}()
		}
	}()
	connection, channels, requests, err := ssh.NewClientConn(connections[0], "synthetic", &ssh.ClientConfig{
		User: "synthetic-user", HostKeyCallback: ssh.FixedHostKey(signer.PublicKey()),
	})
	if err != nil {
		t.Fatal(err)
	}
	client := ssh.NewClient(connection, channels, requests)
	t.Cleanup(func() { client.Close() })
	initial, err := sftp.NewClient(client)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { initial.Close() })
	return &Session{SSH: client, SFTP: initial}
}

func TestFileTransfersUseProvidedReaderAndWriter(t *testing.T) {
	session := syntheticFileSession(t)
	path := filepath.Join(t.TempDir(), "transferred.txt")
	const contents = "synthetic transfer contents"
	if err := session.Put(strings.NewReader(contents), path); err != nil {
		t.Fatal(err)
	}
	session.SFTP.Close()
	var destination bytes.Buffer
	if err := session.Get(&destination, path); err != nil {
		t.Fatal(err)
	}
	session.SFTP.Close()
	if destination.String() != contents {
		t.Fatalf("transferred contents changed: %q", destination.String())
	}
}
