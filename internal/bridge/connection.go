package bridge

import (
	"errors"
	"sync"
	"time"
)

type Pool map[string]*Session
type TunnelPool map[string]*Tunnel

var (
	Connections Pool       = make(Pool)
	Tunnels     TunnelPool = make(TunnelPool)
	mutex       sync.Mutex
)

func (p *Pool) Get(userID string, serverID string) (*Session, error) {
	if conn, ok := Connections[userID+serverID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

func (p *Pool) Set(userID string, serverID string, session *Session) {
	Connections[userID+serverID] = session
}

func (p *Pool) Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(Connections, key)
}

func (t *TunnelPool) Get(remoteHost string, remotePort string, username string) (*Tunnel, error) {
	if tunnel, ok := Tunnels[remoteHost+":"+remotePort+":"+username]; ok {
		tunnel.LastConnection = time.Now()
		return tunnel, nil
	} else {
		return nil, errors.New("tunnel does not exist")
	}
}

func (t *TunnelPool) Set(remoteHost string, remotePort string, username string, tunnel *Tunnel) {
	Tunnels[remoteHost+":"+remotePort+":"+username] = tunnel
}

func (t *TunnelPool) Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(Tunnels, key)
}
