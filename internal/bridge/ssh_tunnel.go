package bridge

import (
	"context"
	"fmt"
	"io"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/avast/retry-go"
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
	password string

	SshClient *ssh.Client

	log logger.Zapper

	ctx    context.Context
	cancel context.CancelFunc

	errHandler func()

	Port           int
	LastConnection time.Time
	Started        bool

	sync.Mutex
}

// Start starts binding to remote end and return error if exists any
func (t *Tunnel) Start() {
	wg := sync.WaitGroup{}
	wg.Add(1)
	go t.bindTunnel(t.ctx, &wg)
	wg.Wait()
}

// Stop collapses tunnel
func (t *Tunnel) Stop() {
	t.Started = false
	t.log.Infow("collapsed tunnel", "details", t)
	if t.SshClient != nil {
		t.SshClient.Close()
	}
	t.cancel()
}

// bindTunnel Binds tunnel with our tunnel object
func (t *Tunnel) bindTunnel(ctx context.Context, wg *sync.WaitGroup) {
	waitDial := sync.WaitGroup{}
	waitDial.Add(1)
	defer waitDial.Done()

	for {
		var once sync.Once // Only print errors once per session
		func() {
			var cl *ssh.Client
			var err error

			// Attempt to dial the remote SSH server.
			err = retry.Do(
				func() error {
					cl, err = ssh.Dial("tcp", t.hostAddr, &ssh.ClientConfig{
						User:            t.user,
						Auth:            t.auth,
						HostKeyCallback: t.hostKeys,
						Timeout:         5 * time.Second,
					})
					if err != nil {
						if strings.Contains(err.Error(), "unable to authenticate") {
							t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
							t.Stop()
							return retry.Unrecoverable(err)
						}
						return err
					}
					return nil
				},
				retry.Attempts(20),
				retry.Delay(1*time.Second),
			)

			if err != nil {
				once.Do(func() {
					t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
					t.Stop()
					t.errHandler()
				})
				return
			}
			defer cl.Close()

			waitDial.Add(1)
			t.SshClient = cl

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
					t.log.Errorw("bind error", "details", fmt.Sprintf("%v, %v", t, err))
					t.Stop()
					t.errHandler()
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
				ln.Close()
			}()

			// Dial once to make sure the connection is established.
			var cn2 net.Conn
			t.Started = false
			err = retry.Do(
				func() error {
					switch t.mode {
					case '>':
						cn2, err = cl.Dial(t.dialType, t.dialAddr)
					case '<':
						cn2, err = net.Dial(t.dialType, t.dialAddr)
					}

					if err != nil {
						if strings.Contains(err.Error(), "open failed") {
							return retry.Unrecoverable(err)
						}
						return err
					}
					cn2.Close()
					return nil
				},
				retry.Attempts(2),
				retry.Delay(1*time.Second),
			)

			if err != nil {
				once.Do(func() {
					t.Started = false
					t.Port = 0
					t.Stop()
					t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
					t.errHandler()
				})
				return
			}
			// Dial ends

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
						t.log.Errorw("accept error", "details", fmt.Sprintf("%v, %v", t, err))
						t.Stop()
						t.errHandler()
					})
					return
				}
				waitDial.Add(1)

				// The inbound connection is established. Make sure we close it eventually.
				go t.dialTunnel(bindCtx, &waitDial, cl, cn1)
			}
		}()

		select {
		case <-ctx.Done():
			return
		case <-time.After(30 * time.Second):
			fmt.Printf("(%v) retrying...\n", t)
		}
	}
}

// dialTunnel dials connection and waits until context get cancelled
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
	var once sync.Once
	var cn2 net.Conn
	var err error

	err = retry.Do(
		func() error {
			switch t.mode {
			case '>':
				cn2, err = client.Dial(t.dialType, t.dialAddr)
			case '<':
				cn2, err = net.Dial(t.dialType, t.dialAddr)
			}

			if err != nil {
				if strings.Contains(err.Error(), "open failed") {
					return retry.Unrecoverable(err)
				}
				return err
			}
			return nil
		},
		retry.Attempts(2),
		retry.Delay(1*time.Second),
	)
	if err != nil {
		once.Do(func() {
			t.Stop()
			t.log.Errorw("ssh dial error", "details", fmt.Sprintf("%v, %v", t, err))
			t.errHandler()
		})
		return
	}

	go func() {
		<-connCtx.Done()
		cn2.Close()
	}()

	// Copy bytes from one connection to the other until one side closes.
	var waitDial sync.WaitGroup
	waitDial.Add(2)
	go func() {
		defer waitDial.Done()
		defer cancel()
		if _, err := io.Copy(cn1, cn2); err != nil {
			once.Do(func() {
				t.Stop()
				t.errHandler()
				t.log.Errorw("connection error", "details", fmt.Sprintf("%v, %v", t, err))
			})
		}
	}()
	go func() {
		defer waitDial.Done()
		defer cancel()
		if _, err := io.Copy(cn2, cn1); err != nil {
			once.Do(func() {
				t.Stop()
				t.errHandler()
				t.log.Errorw("connection error", "details", fmt.Sprintf("%v, %v", t, err))
			})
		}
	}()
	waitDial.Wait()
}

// String returns readable format of tunnel struct
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

// CreateTunnel starts a new tunnel instance and sets it into TunnelPool
func CreateTunnel(remoteHost, remotePort, username, password, sshPort string) int {
	// Creating a tunnel cannot exceed 25 seconds
	ch := make(chan int)
	time.AfterFunc(25*time.Second, func() {
		ch <- 1
	})

	// Check if a tunnel exists with this remoteHost, remotePort and username
	t, err := Tunnels.Get(remoteHost, remotePort, username)

	// Check if existing tunnel started, if not wait until starts (max: 25sec)
	if err == nil {
		if t.password != password {
			return 0
		}

	startedLoop:
		for {
			if t.Started {
				break
			}

			select {
			case <-ch:
				break startedLoop
			case <-t.ctx.Done():
				break startedLoop
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}

		t.LastConnection = time.Now()
		return t.Port
	}

	// This part from now creates a new tunnel
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

	ctx, cancel := context.WithCancel(context.Background())
	sshTunnel := &Tunnel{
		auth:     []ssh.AuthMethod{ssh.RetryableAuthMethod(ssh.Password(password), 3)},
		hostKeys: ssh.InsecureIgnoreHostKey(),
		user:     username,
		mode:     '>',
		hostAddr: net.JoinHostPort(remoteHost, sshPort),
		dialAddr: dial,
		dialType: dialType,
		bindAddr: net.JoinHostPort("127.0.0.1", fmt.Sprintf("%d", port)),
		log:      logger.Sugar(),
		errHandler: func() {
			Tunnels.Delete(remoteHost + ":" + remotePort + ":" + username)
		},
		password:       password,
		Port:           port,
		LastConnection: time.Now(),
		Started:        false,
		ctx:            ctx,
		cancel:         cancel,
	}

	Tunnels.Set(remoteHost, remotePort, username, sshTunnel)
	go sshTunnel.Start()

loop:
	for {
		if sshTunnel.Started {
			break
		}

		select {
		case <-ch:
			break loop
		case <-sshTunnel.ctx.Done():
			break loop
		default:
			time.Sleep(10 * time.Millisecond)
			continue
		}
	}

	if !sshTunnel.Started {
		cancel()
		return 0
	}

	return sshTunnel.Port
}
