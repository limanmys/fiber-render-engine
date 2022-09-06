package bridge

import (
	"net"
	"time"

	"github.com/limanmys/render-engine/pkg/helpers"
)

func Clean() {
	now := time.Now()
	for key, session := range Connections {
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
				session.SSH.SendRequest("keepalive@liman.dev", true, nil)
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
	time.Sleep(5 * time.Second)
	go Clean()
}

func closeSession(s *Session, key string) {
	s.Mutex.Lock()
	s.CloseAllConnections()
	s.Mutex.Unlock()

	Connections.Delete(key)
}
