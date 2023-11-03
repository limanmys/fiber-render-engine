package process_queue

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/bridge"
	"github.com/limanmys/render-engine/internal/file"
	"github.com/limanmys/render-engine/internal/liman"
	"gorm.io/gorm"
)

type InstallPackage struct {
	Queue  *models.Queue
	DB     *gorm.DB
	UserID string
}

func (p InstallPackage) Process() error {
	p.Queue.UpdateStatus(models.StatusProcessing)

	location, err := p.sendPackageToRemoteServer()
	if err != nil {
		p.Queue.UpdateError(err.Error())
		return err
	}

	err = p.installPackage(location)
	if err != nil {
		p.Queue.UpdateError(err.Error())
		return err
	}

	p.Queue.UpdateStatus(models.StatusDone)

	return nil
}

func (p InstallPackage) sendPackageToRemoteServer() (string, error) {
	path := p.Queue.Data["path"].(string)
	fileName := filepath.Base(path)

	location, err := file.PutFileHandler(
		p.UserID,
		p.Queue.Data["server_id"].(string),
		"/tmp/"+fileName,
		path,
	)

	if err == nil {
		os.Remove(path)
	}

	return location, err
}

func (p InstallPackage) installPackage(location string) error {
	server, err := liman.GetServer(&models.Server{ID: p.Queue.Data["server_id"].(string)})
	if err != nil {
		return err
	}

	shell, err := bridge.GetSession(
		p.UserID,
		server.ID,
		server.IPAddress,
	)
	if err != nil {
		return err
	}

	osType, _ := shell.Run("cat /etc/os-release | grep ^ID_LIKE | cut -d '=' -f2 | xargs")
	packageManager := ""
	if strings.ToLower(osType) == "debian" {
		packageManager = "apt"
	} else {
		packageManager = "yum"
	}

	switch packageManager {
	case "apt":
		_, err := shell.Run("DEBIAN_FRONTEND=noninteractive sudo -p liman-pass-sudo apt install -fyq " + location)
		if err != nil {
			return err
		}
	case "yum":
		_, err := shell.Run("sudo -p liman-pass-sudo /usr/bin/yum install -yq " + location)
		if err != nil {
			return err
		}
	}
	shell.Run("rm -rf " + location)

	return nil
}
