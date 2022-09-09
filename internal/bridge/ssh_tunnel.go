package bridge

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/phayes/freeport"
	"golang.org/x/crypto/ssh"
)

type Tunnel struct {
	auth     []ssh.AuthMethod
	hostKeys ssh.HostKeyCallback
	mode     byte // '>' for forward, '<' for reverse
	user     string
	hostAddr string
	bindAddr string
	dialAddr string
	dialType string

	SshClient *ssh.Client

	log logger.Zapper

	ctx    context.Context
	cancel context.CancelFunc

	errHandler func()

	Port           int
	LastConnection time.Time
	Mutex          sync.Mutex
	Started        bool
}

var mut sync.Mutex = sync.Mutex{}

func CreateTunnel(remoteHost, remotePort, username, password string) int {
	t, err := Tunnels.Get(remoteHost, remotePort, username)
	if err == nil {
		for {
			if t.Started {
				break
			}
		}

		t.LastConnection = time.Now()
		return t.Port
	}

	mut.Lock()
	defer mut.Unlock()

	port, err := freeport.GetFreePort()
	if err != nil {
		logger.Sugar().Errorw(err.Error())
		return 0
	}

	dial := net.JoinHostPort("127.0.0.1", remotePort)
	dialType := "tcp"

	if _, err := strconv.Atoi(remotePort); err != nil {
		dial = remotePort
		dialType = "unix"
	}

	sshTunnel := &Tunnel{
		auth:     []ssh.AuthMethod{ssh.Password(password)},
		hostKeys: ssh.InsecureIgnoreHostKey(),
		user:     username,
		mode:     '>',
		hostAddr: net.JoinHostPort(remoteHost, "22"),
		dialAddr: dial,
		dialType: dialType,
		bindAddr: net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)),
		log:      logger.Sugar(),
		errHandler: func() {
			Tunnels.Delete(remoteHost + ":" + remotePort + ":" + username)
		},
		Port:           port,
		LastConnection: time.Now(),
		Started:        false,
	}

	Tunnels.Set(remoteHost, remotePort, username, sshTunnel)

	hasError := sshTunnel.Start()
	if !hasError {
		for {
			if sshTunnel.Started {
				break
			}
		}

		return port
	}

	Tunnels.Delete(remoteHost + ":" + remotePort + ":" + username)
	return 0
}

func (t *Tunnel) String() string {
	var left, right string
	mode := "<?>"
	switch t.mode {
	case '>':
		left, mode, right = t.bindAddr, "->", t.dialAddr
	case '<':
		left, mode, right = t.dialAddr, "<-", t.bindAddr
	}
	return fmt.Sprintf("%s@%s | %s %s %s", t.user, t.hostAddr, left, mode, right)
}

func (t *Tunnel) Start() bool {
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	hasError := false
	wg := sync.WaitGroup{}
	wg.Add(1)
	go t.bindTunnel(ctx, &wg, &hasError)
	wg.Wait()

	return hasError
}

func (t *Tunnel) Stop() {
	t.log.Infow("collapsed tunnel", "details", t)
	t.cancel()
}

func (t *Tunnel) bindTunnel(ctx context.Context, wg *sync.WaitGroup, hasError *bool) {
	wgt := sync.WaitGroup{}
	wgt.Add(1)
	defer wgt.Done()
	for {
		var once sync.Once // Only print errors once per session
		func() {
			// Connect to the server host via SSH.
			cl, err := ssh.Dial("tcp", t.hostAddr, &ssh.ClientConfig{
				User:            t.user,
				Auth:            t.auth,
				HostKeyCallback: t.hostKeys,
				Timeout:         5 * time.Second,
			})
			if err != nil {
				once.Do(func() {
					t.errHandler()
					t.Stop()
					*hasError = true
					t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
					wg.Done()
				})
				return
			}
			wgt.Add(1)
			t.SshClient = cl

			defer cl.Close()

			// Attempt to bind to the inbound socket.
			var ln net.Listener
			switch t.mode {
			case '>':
				ln, err = net.Listen("tcp", t.bindAddr)
			case '<':
				ln, err = cl.Listen("tcp", t.bindAddr)
			}
			if err != nil {
				once.Do(func() {
					t.errHandler()
					t.log.Errorw("bind error", "details", fmt.Sprintf("%v, %v", t, err))
					t.Stop()
					*hasError = true
					wg.Done()
				})
				return
			}

			// The socket is binded. Make sure we close it eventually.
			bindCtx, cancel := context.WithCancel(ctx)
			defer cancel()
			go func() {
				cl.Wait()
				cancel()
			}()
			go func() {
				<-bindCtx.Done()
				once.Do(func() {}) // Suppress future errors
				ln.Close()
			}()

			t.Started = true
			t.log.Infow("binded tunnel", "details", t)
			wg.Done()
			defer t.log.Infow("collapsed tunnel", "details", t)
			defer t.errHandler()
			// Accept all incoming connections.
			for {
				cn1, err := ln.Accept()
				if err != nil {
					once.Do(func() {
						t.Stop()
						*hasError = true
						t.errHandler()
						t.log.Errorw("accept error", "details", fmt.Sprintf("%v, %v", t, err))
						wg.Done()
					})
					return
				}
				wgt.Add(1)

				go t.dialTunnel(bindCtx, &wgt, cl, cn1, hasError)
			}
		}()

		select {
		case <-ctx.Done():
			return
		}
	}
}

func (t *Tunnel) dialTunnel(ctx context.Context, wg *sync.WaitGroup, client *ssh.Client, cn1 net.Conn, hasError *bool) {
	defer wg.Done()

	// The inbound connection is established. Make sure we close it eventually.
	connCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-connCtx.Done()
		cn1.Close()
	}()

	// Establish the outbound connection.
	var cn2 net.Conn
	var err error
	switch t.mode {
	case '>':
		cn2, err = client.Dial(t.dialType, t.dialAddr)
	case '<':
		cn2, err = net.Dial(t.dialType, t.dialAddr)
	}
	if err != nil {
		t.Stop()
		t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
		t.errHandler()
		*hasError = true

		wg.Done()
		return
	}

	go func() {
		<-connCtx.Done()
		cn2.Close()
	}()

	t.log.Infow("connection established", "details", t)
	defer t.log.Infow("connection closed", "details", t)

	// Copy bytes from one connection to the other until one side closes.
	var once sync.Once
	var wg2 sync.WaitGroup
	wg2.Add(2)
	go func() {
		defer wg2.Done()
		defer cancel()
		if _, err := io.Copy(cn1, cn2); err != nil {
			once.Do(func() {
				t.errHandler()
				t.log.Errorw("connection error", "details", fmt.Sprintf("%v, %v", t, err))
			})
		}
		once.Do(func() {}) // Suppress future errors
	}()
	go func() {
		defer wg2.Done()
		defer cancel()
		if _, err := io.Copy(cn2, cn1); err != nil {
			once.Do(func() {
				t.errHandler()
				t.log.Errorw("connection error", "details", fmt.Sprintf("%v, %v", t, err))
			})
		}
		once.Do(func() {}) // Suppress future errors
	}()
	wg2.Wait()
}
