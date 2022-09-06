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

	retryInterval time.Duration

	log logger.Zapper

	ctx    context.Context
	cancel context.CancelFunc

	errHandler func()

	Port           int
	LastConnection time.Time
	Mutex          sync.Mutex
}

var mut sync.Mutex = sync.Mutex{}

func CreateTunnel(remoteHost string, remotePort string, username string, password string) int {
	t, err := Tunnels.Get(remoteHost, remotePort, username)
	if err == nil {
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
		auth:          []ssh.AuthMethod{ssh.Password(password)},
		hostKeys:      ssh.InsecureIgnoreHostKey(),
		user:          username,
		mode:          '>',
		hostAddr:      net.JoinHostPort(remoteHost, "22"),
		dialAddr:      dial,
		dialType:      dialType,
		bindAddr:      net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)),
		retryInterval: 30 * time.Second,
		log:           logger.Sugar(),
		errHandler: func() {
			Tunnels.Delete(remoteHost + ":" + remotePort + ":" + username)
		},
		Port:           port,
		LastConnection: time.Now(),
	}
	sshTunnel.Start()
	Tunnels.Set(remoteHost, remotePort, username, sshTunnel)

	return port
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

func (t *Tunnel) Start() {
	ctx, cancel := context.WithCancel(context.Background())
	t.ctx = ctx
	t.cancel = cancel
	wg := sync.WaitGroup{}
	wg.Add(1)
	go t.bindTunnel(ctx, &wg)
	wg.Wait()
}

func (t *Tunnel) Stop() {
	t.cancel()
}

func (t *Tunnel) bindTunnel(ctx context.Context, wg *sync.WaitGroup) {
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
					t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
				})
				return
			}
			wgt.Add(1)
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

			t.log.Infow("binded tunnel", "details", t)
			wg.Done()
			defer t.log.Infow("collapsed tunnel", "details", t)
			defer t.errHandler()
			// Accept all incoming connections.
			for {
				cn1, err := ln.Accept()
				if err != nil {
					once.Do(func() {
						t.errHandler()
						t.log.Errorw("accept error", "details", fmt.Sprintf("%v, %v", t, err))
					})
					return
				}
				wgt.Add(1)
				go t.dialTunnel(bindCtx, &wgt, cl, cn1)
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-time.After(t.retryInterval):
			t.log.Infow("retrying...", "details", t)
		}
	}
}

func (t *Tunnel) dialTunnel(ctx context.Context, wg *sync.WaitGroup, client *ssh.Client, cn1 net.Conn) {
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
		t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
		t.errHandler()
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
