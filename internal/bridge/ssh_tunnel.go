package bridge

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/phayes/freeport"
	"golang.org/x/crypto/ssh"
)

type Tunnel struct {
	auth        []ssh.AuthMethod
	hostKeys    ssh.HostKeyCallback
	mode        byte // '>' for forward, '<' for reverse
	user        string
	hostAddr    string
	bindAddr    string
	dialAddr    string
	dialType    string
	password    string
	SshClient   *ssh.Client
	log         logger.Zapper
	ctx         context.Context
	cancel      context.CancelFunc
	errHandler  func()
	Port        int
	LastConnection time.Time
	Started     bool
	sync.Mutex
}

// Start initiates the tunnel binding process
func (t *Tunnel) Start() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go t.bindTunnel(t.ctx, &wg)
	wg.Wait()
}

// Stop terminates the tunnel and closes connections
func (t *Tunnel) Stop() {
	t.Started = false
	t.log.Infow("collapsed tunnel", "details", t)
	if t.SshClient != nil {
		t.SshClient.Close()
	}
	t.cancel()
}

// bindTunnel establishes the SSH tunnel
func (t *Tunnel) bindTunnel(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()
	defer t.Stop()

	for {
		var once sync.Once
		cl, err := t.dialSSH()
		if err != nil {
			once.Do(func() {
				t.log.Errorw("SSH dial error", "details", fmt.Sprintf("%v, %v", t, err))
				t.errHandler()
			})
			return
		}
		defer cl.Close()

		ln, err := t.bindSocket(cl)
		if err != nil {
			once.Do(func() {
				t.log.Errorw("Bind error", "details", fmt.Sprintf("%v, %v", t, err))
				t.errHandler()
			})
			return
		}
		defer ln.Close()

		t.Started = true
		wg.Done()
		t.log.Infow("Tunnel bound", "details", t)

		// Handle incoming connections
		for {
			conn, err := ln.Accept()
			if err != nil {
				once.Do(func() {
					t.log.Errorw("Accept error", "details", fmt.Sprintf("%v, %v", t, err))
					t.errHandler()
				})
				return
			}
			go t.dialTunnel(ctx, cl, conn)
		}
	}
}

// dialSSH attempts to create an SSH connection
func (t *Tunnel) dialSSH() (*ssh.Client, error) {
	var cl *ssh.Client
	err := retry.Do(
		func() error {
			var err error
			cl, err = ssh.Dial("tcp", t.hostAddr, &ssh.ClientConfig{
				User:            t.user,
				Auth:            t.auth,
				HostKeyCallback: t.hostKeys,
				Timeout:         5 * time.Second,
			})
			if err != nil && strings.Contains(err.Error(), "unable to authenticate") {
				return retry.Unrecoverable(err)
			}
			return err
		},
		retry.Attempts(20),
		retry.Delay(1*time.Second),
	)
	return cl, err
}

// bindSocket binds the local or remote listener based on the mode
func (t *Tunnel) bindSocket(cl *ssh.Client) (net.Listener, error) {
	switch t.mode {
	case '>':
		return net.Listen("tcp", t.bindAddr)
	case '<':
		return cl.Listen("tcp", t.bindAddr)
	default:
		return nil, errors.New("invalid mode")
	}
}

// dialTunnel handles the connection from local to remote via SSH tunnel
func (t *Tunnel) dialTunnel(ctx context.Context, client *ssh.Client, conn net.Conn) {
	defer conn.Close()

	var remoteConn net.Conn
	var err error
	err = retry.Do(
		func() error {
			switch t.mode {
			case '>':
				remoteConn, err = client.Dial(t.dialType, t.dialAddr)
			case '<':
				remoteConn, err = net.Dial(t.dialType, t.dialAddr)
			}
			return err
		},
		retry.Attempts(2),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		t.log.Errorw("Connection error", "details", fmt.Sprintf("%v, %v", t, err))
		t.errHandler()
		return
	}
	defer remoteConn.Close()

	// Copy data between local and remote connections
	go io.Copy(conn, remoteConn)
	go io.Copy(remoteConn, conn)
}

// CreateTunnelInternalKey initiates a tunnel based on request data
func CreateTunnelInternalKey(c *fiber.Ctx) (int, error) {
	credentials, err := liman.GetCredentials(&models.User{
		ID: c.Locals("user_id").(string),
	}, &models.Server{
		ID: c.FormValue("server_id"),
	})
	if err != nil {
		return 0, err
	}

	port, err := CreateTunnel(
		c.FormValue("remote_host"),
		c.FormValue("remote_port"),
		credentials.Username,
		credentials.Key,
		credentials.Port,
		credentials.Type,
	)
	return port, err
}

// CreateTunnel sets up a new SSH tunnel instance
func CreateTunnel(remoteHost, remotePort, username, password, sshPort, connType string) (int, error) {
	ch := make(chan struct{})
	defer close(ch)
	time.AfterFunc(25*time.Second, func() { ch <- struct{}{} })

	t, err := Tunnels.Get(remoteHost, remotePort, username)
	if err == nil && t.password == password {
		select {
		case <-ch:
			return 0, errors.New("tunnel setup timeout")
		case <-t.ctx.Done():
			t.LastConnection = time.Now()
			return t.Port, nil
		}
	}

	port, err := freeport.GetFreePort()
	if err != nil {
		return 0, errors.New("failed to find a free port")
	}

	dialAddr := net.JoinHostPort("127.0.0.1", remotePort)
	if _, err := strconv.Atoi(remotePort); err != nil {
		dialAddr = remotePort
	}

	ctx, cancel := context.WithCancel(context.Background())
	sshTunnel := &Tunnel{
		hostKeys: ssh.InsecureIgnoreHostKey(),
		user:     username,
		mode:     '>',
		hostAddr: net.JoinHostPort(remoteHost, sshPort),
		dialAddr: dialAddr,
		dialType: "tcp",
		bindAddr: net.JoinHostPort("127.0.0.1", strconv.Itoa(port)),
		log:      logger.Sugar(),
		password: password,
		Port:     port,
		ctx:      ctx,
		cancel:   cancel,
		errHandler: func() {
			Tunnels.Delete(remoteHost + ":" + remotePort + ":" + username)
		},
	}

	if connType == "ssh" {
		sshTunnel.auth = []ssh.AuthMethod{ssh.Password(password)}
	}

	Tunnels.Set(remoteHost, remotePort, username, sshTunnel)
	go sshTunnel.Start()

	select {
	case <-ch:
		cancel()
		return 0, errors.New("tunnel setup timeout")
	case <-sshTunnel.ctx.Done():
		return sshTunnel.Port, nil
	}
}
