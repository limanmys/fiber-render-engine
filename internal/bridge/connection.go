package bridge

import (
	"errors"
	"strings"
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

// Get connection from pool
func (p *Pool) Get(userID, serverID string) (*Session, error) {
	if conn, ok := Connections[userID+serverID]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// Set connection on pool
func (p *Pool) Set(userID, serverID string, session *Session) {
	Connections[userID+serverID] = session
}

// GetRaw connection object from pool
func (p *Pool) GetRaw(userID, remoteHost, username string) (*Session, error) {
	if conn, ok := Connections[userID+remoteHost+username]; ok {
		conn.LastConnection = time.Now()
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// SetRaw connection object on pool
func (p *Pool) SetRaw(userID, remoteHost, username string, session *Session) {
	Connections[userID+remoteHost+username] = session
}

// Delete connection object from pool
func (p *Pool) Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(Connections, key)
}

// Get Tunnel connection from pool
func (t *TunnelPool) Get(remoteHost, remotePort, username string) (*Tunnel, error) {
	if tunnel, ok := Tunnels[remoteHost+":"+remotePort+":"+username]; ok {
		tunnel.LastConnection = time.Now()
		return tunnel, nil
	} else {
		return nil, errors.New("tunnel does not exist")
	}
}

// Set Tunnel connection to pool
func (t *TunnelPool) Set(remoteHost, remotePort, username string, tunnel *Tunnel) {
	Tunnels[remoteHost+":"+remotePort+":"+username] = tunnel
}

// Delete Tunnel connection from pool
func (t *TunnelPool) Delete(key string) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(Tunnels, key)
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
