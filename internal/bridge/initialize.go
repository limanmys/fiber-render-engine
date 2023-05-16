package bridge

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

// GetSession gets session from memory and if does not exists creates a new one
func GetSession(userID, serverID, host string) (*Session, error) {
	session, err := Sessions.Get(userID, serverID)
	if err != nil {
		session = &Session{}
		isConnected := session.CreateShell(userID, serverID, host)
		if !isConnected {
			return nil, logger.FiberError(fiber.StatusForbidden, "cannot connect to server")
		}
		Sessions.Set(userID, serverID, session)
	}

	session.LastConnection = time.Now()
	return session, nil
}

// CreateShell creates a new shell with credential type and sets it on session
func (s *Session) CreateShell(userID, serverID, host string) bool {
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
	} else if credentials.Type == "ssh_certificate" {
		conn, err := InitShellWithCert(s.Username, s.password, s.IpAddr, s.Port)
		if err != nil {
			return false
		}
		s.SSH = conn
	} else if strings.Contains(credentials.Type, "winrm") {
		secure := true
		if strings.Contains(credentials.Type, "insecure") {
			secure = false
		}

		conn, err := InitWinRm(s.Username, s.password, s.IpAddr, s.Port, secure)
		if err != nil {
			return false
		}
		s.WinRM = conn

		winrmPath, err := s.Run("echo $env:TEMP")
		if err != nil {
			return false
		}

		s.WindowsPath = strings.TrimSpace(winrmPath) + "\\"
		s.WindowsLetter = s.WindowsPath[0:1]
		s.WindowsPath = s.WindowsPath[3:]
	} else {
		return false
	}

	return true
}

// CreateRaw creates a new raw shell with credential type and sets it on session
func (s *Session) CreateRaw(connectionType, username, password, host, port string) bool {
	s.Username = username
	s.password = password
	s.IpAddr = host
	s.Port = port
	s.LastConnection = time.Now()

	if connectionType == "ssh" {
		conn, err := InitShellWithPassword(s.Username, s.password, s.IpAddr, s.Port)
		if err != nil {
			return false
		}
		s.SSH = conn
	} else if connectionType == "ssh_certificate" {
		conn, err := InitShellWithCert(s.Username, s.password, s.IpAddr, s.Port)
		if err != nil {
			return false
		}
		s.SSH = conn
	} else if strings.Contains(connectionType, "winrm") {
		secure := true
		if strings.Contains(connectionType, "insecure") {
			secure = false
		}

		conn, err := InitWinRm(s.Username, s.password, s.IpAddr, s.Port, secure)
		if err != nil {
			return false
		}
		s.WinRM = conn

		winrmPath, err := s.Run("echo $env:TEMP")
		if err != nil {
			return false
		}

		s.WindowsPath = strings.TrimSpace(winrmPath) + "\\"
		s.WindowsLetter = s.WindowsPath[0:1]
		s.WindowsPath = s.WindowsPath[3:]
	} else {
		return false
	}

	return true
}

// CreateFileSesssion creates a new file session with credential type and sets it on session
func (s *Session) CreateFileConnection(userID, serverID, host string) bool {
	credentials, err := liman.GetCredentials(&models.User{ID: userID}, &models.Server{ID: serverID})
	if err != nil {
		return false
	}

	s.Username = credentials.Username
	s.password = credentials.Key
	s.IpAddr = host
	s.Port = credentials.Port
	s.LastConnection = time.Now()

	if credentials.Type == "ssh" || credentials.Type == "ssh_certificate" {
		if s.SFTP != nil {
			return true
		}

		flag := s.CreateShell(userID, serverID, host)
		if !flag {
			return false
		}
		s.SFTP = OpenSFTPConnection(s.SSH)
	} else if strings.Contains(credentials.Type, "winrm") {
		if s.SMB != nil {
			return true
		}

		smb, err := OpenSmbConnection(s.IpAddr, s.Username, s.password)
		if err != nil {
			return false
		}

		s.SMB = smb
	} else {
		return false
	}

	return true
}
