package bridge

import (
	"net"
	"time"

	"github.com/limanmys/render-engine/pkg/helpers"
	"github.com/limanmys/render-engine/pkg/logger"
)

// Clean clears long standing sessions from memory
func Clean() {
	now := time.Now()
	for key, session := range GetConnections().connections {
		if now.Sub(session.LastConnection).Seconds() > 266 {
			closeSession(session, key)
			continue
		}

		if session.SSH == nil {
			continue
		}

		if session.IpAddr != "" && session.Port != "" {
			ip, err := helpers.ResolveIP(session.IpAddr)
			if err != nil {
				closeSession(session, key)
				continue
			}

			_, err = net.DialTimeout(
				"tcp",
				net.JoinHostPort(ip, session.Port),
				10*time.Second,
			)
			if err != nil {
				closeSession(session, key)
				continue
			}
		}

		ch := make(chan int, 1)
		go func() {
			select {
			case <-time.After(10 * time.Second):
			case <-ch:
				return
			default:
				_, _, err := session.SSH.SendRequest("keepalive@liman.dev", true, nil)
				if err != nil {
					closeSession(session, key)
					logger.Sugar().Warnw("error when sending request")
				}

				ch <- 1
			}
		}()

		select {
		case <-ch:
			continue
		case <-time.After(10 * time.Second):
			closeSession(session, key)
			continue
		}
	}

	for key, tunnel := range GetTunnels().connections {
		if now.Sub(tunnel.LastConnection).Seconds() > 266 {
			closeTunnel(tunnel, key)
			continue
		}

		if tunnel.SshClient == nil {
			continue
		}

		_, err := net.DialTimeout(
			"tcp",
			tunnel.hostAddr,
			10*time.Second,
		)
		if err != nil {
			closeTunnel(tunnel, key)
			continue
		}

		ch := make(chan int, 1)
		go func() {
			select {
			case <-time.After(10 * time.Second):
			case <-ch:
				return
			default:
				_, _, err := tunnel.SshClient.SendRequest("keepalive@liman.dev", true, nil)
				if err != nil {
					closeTunnel(tunnel, key)
					logger.Sugar().Warnw("error when sending request")
				}

				ch <- 1
			}
		}()

		select {
		case <-ch:
			continue
		case <-time.After(10 * time.Second):
			closeTunnel(tunnel, key)
			continue
		}
	}

	time.Sleep(10 * time.Second)
	go Clean()
}

func closeSession(s *Session, key string) {
	defer s.Unlock()
	s.Lock()
	s.CloseAllConnections()

	GetConnections().Delete(key)
}

func closeTunnel(t *Tunnel, key string) {
	defer t.Unlock()
	t.Lock()
	t.Stop()

	GetTunnels().Delete(key)
}
