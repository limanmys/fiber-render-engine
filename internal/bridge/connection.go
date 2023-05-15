package bridge

import (
	"errors"
	"strings"
	"sync"
)

type poolable interface {
	Session | Tunnel
}

type pool[T poolable] struct {
	connections map[string]*T
	sync.Mutex
}

var (
	connections *pool[Session] = &pool[Session]{connections: make(map[string]*Session)}
	tunnels     *pool[Tunnel]  = &pool[Tunnel]{connections: make(map[string]*Tunnel)}
)

// Get connections object
func GetConnections() *pool[Session] {
	return connections
}

// Get tunnels object
func GetTunnels() *pool[Tunnel] {
	return tunnels
}

// Get connection from pool
func (p *pool[T]) Get(key string) (*T, error) {
	if conn, ok := p.connections[key]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// Set connection on pool
func (p *pool[T]) Set(key string, session *T) {
	defer p.Unlock()
	p.Lock()

	p.connections[key] = session
}

// Delete connection object from pool
func (p *pool[T]) Delete(key string) {
	defer p.Unlock()
	p.Lock()

	delete(p.connections, key)
}

// GetRaw connection object from pool
func (p *pool[T]) GetRaw(key string) (*T, error) {
	if conn, ok := p.connections[key]; ok {
		return conn, nil
	} else {
		return nil, errors.New("connection does not exist")
	}
}

// SetRaw connection object on pool
func (p *pool[T]) SetRaw(key string, session *T) {
	defer p.Unlock()
	p.Lock()

	p.connections[key] = session
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
