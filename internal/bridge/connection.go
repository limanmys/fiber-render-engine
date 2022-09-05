package bridge

import (
	"errors"
	"sync"
)

type Pool map[string]*Session

var (
	Connections Pool = make(Pool)
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
