package bridge

import (
	"errors"
	"strings"
	"sync"
	"time"
)

type Pool struct {
	Connections map[string]*Session
	sync.Mutex
}
type TunnelPool struct {
	Connections map[string]*Tunnel
	sync.Mutex
}

var (
	Sessions Pool       = Pool{Connections: make(map[string]*Session)}
	Tunnels  TunnelPool = TunnelPool{Connections: make(map[string]*Tunnel)}
)

// Get connection from pool
func (p *Pool) Get(userID, serverID string) (*Session, error) {
	if conn, ok := p.Connections[userID+serverID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// Set connection on pool
func (p *Pool) Set(userID, serverID string, session *Session) {
	p.Lock()
	defer p.Unlock()
	p.Connections[userID+serverID] = session
}

// GetRaw connection object from pool
func (p *Pool) GetRaw(userID, remoteHost, username string) (*Session, error) {
	if conn, ok := p.Connections[userID+remoteHost+username]; ok {
		conn.LastConnection = time.Now()
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// SetRaw connection object on pool
func (p *Pool) SetRaw(userID, remoteHost, username string, session *Session) {
	p.Lock()
	defer p.Unlock()
	p.Connections[userID+remoteHost+username] = session
}

// Delete connection object from pool
func (p *Pool) Delete(key string) {
	p.Lock()
	defer p.Unlock()
	delete(p.Connections, key)
}

// Get Tunnel connection from pool
func (t *TunnelPool) Get(remoteHost, remotePort, username string) (*Tunnel, error) {
	if tunnel, ok := Tunnels.Connections[remoteHost+":"+remotePort+":"+username]; ok {
		tunnel.LastConnection = time.Now()
		return tunnel, nil
	} else {
		return nil, errors.New("tunnel does not exist")
	}
}

// Set Tunnel connection to pool
func (t *TunnelPool) Set(remoteHost, remotePort, username string, tunnel *Tunnel) {
	t.Lock()
	defer t.Unlock()
	Tunnels.Connections[remoteHost+":"+remotePort+":"+username] = tunnel
}

// Delete Tunnel connection from pool
func (t *TunnelPool) Delete(key string) {
	t.Lock()
	defer t.Unlock()
	delete(Tunnels.Connections, key)
}

// VerifyAuth verifies key when connecting
func VerifyAuth(username, password, ipAddress, port, keyType string) bool {
	if keyType == "ssh" {
		return VerifySSH(username, password, ipAddress, port)
	} else if keyType == "ssh_certificate" {
		return VerifySSHCertificate(username, password, ipAddress, port)
	} else if strings.Contains(keyType, "winrm") {
		secure := true
		if strings.Contains(keyType, "insecure") {
			secure = false
		}

		return VerifyWinRm(username, password, ipAddress, port, secure)
	}
	return true
}
