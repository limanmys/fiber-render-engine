package file

import (
	"io"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/liman"
	"github.com/limanmys/render-engine/pkg/logger"
)

func PutFileHandler(user_id, server_id, remote_path string, localFile io.Reader) (string, error) {
	server, err := liman.GetServer(&models.Server{ID: server_id})
	if err != nil {
		return "", err
	}

	session, err := bridge.GetSession(
		user_id,
		server.ID,
		server.IPAddress,
	)
	if err != nil {
		return "", err
	}

	established := session.CreateFileConnection(
		user_id,
		server.ID,
		server.IPAddress,
	)
	if !established {
		return "", logger.FiberError(fiber.StatusServiceUnavailable, "cannot establish file connection")
	}

	remotePath, err := RemotePath(server.Os, session.WindowsPath, remote_path)
	if err != nil {
		return "", err
	}

	err = session.Put(localFile, remotePath)
	if err != nil {
		return "", err
	}
	return remotePath, nil
}
