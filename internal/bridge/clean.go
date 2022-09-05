package bridge

import (
	"time"
)

func Clean() {
	now := time.Now()
	for key, session := range Connections {
		if now.Sub(session.LastConnection).Seconds() > 266 {
			session.Mutex.Lock()
			session.CloseAllConnections()
			session.Mutex.Unlock()

			Connections.Delete(key)
			continue
		}
	}
	time.Sleep(time.Second)
	go Clean()
}
