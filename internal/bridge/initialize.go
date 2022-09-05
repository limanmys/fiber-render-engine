package bridge

import (
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

func GetConnection(userID string, serverID string, host string) (*Session, error) {
	session, err := Connections.Get(userID, serverID)
	if err != nil {
		session = &Session{}
		isConnected := session.CreateShell(userID, serverID, host)
		if !isConnected {
			return nil, logger.FiberError(fiber.StatusForbidden, "cannot connect to server")
		}
		Connections.Set(userID, serverID, session)
	}

	session.LastConnection = time.Now()
	return session, nil
}

func (s *Session) CreateShell(userID string, serverID string, host string) bool {
	credentials, err := liman.GetCredentials(&models.User{ID: userID}, &models.Server{ID: serverID})
	if err != nil {
		return false
	}

	s.Username = credentials.Username
	s.password = credentials.Key
	s.IpAddr = host
	s.Port = credentials.Port
	s.LastConnection = time.Now()

	if credentials.Type == "ssh" {
		conn, err := InitShellWithPassword(s.Username, s.password, s.IpAddr, s.Port)
		if err != nil {
			return false
		}
		s.SSH = conn
	} else {
		return false
	}

	return true
}
