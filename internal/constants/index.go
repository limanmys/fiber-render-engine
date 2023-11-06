package constants

import (
	"time"

	"github.com/go-co-op/gocron"
)

const (
	LIMAN_PATH            = "/liman"
	CORE_PATH             = LIMAN_PATH + "/server"
	EXEC_RUNNER           = "/bin/bash"
	EXTENSIONS_PATH       = LIMAN_PATH + "/extensions"
	FUNCTIONS_FILE_PATH   = "/views/functions.php"
	EXTENSION_PUBLIC_PATH = "%s/eklenti/%s/public/"
	NAVIGATION_ROUTE      = "/l/%s/%s"
	SANDBOX_PATH          = LIMAN_PATH + "/sandbox/php/index.php"
	KEYS_PATH             = LIMAN_PATH + "/keys"
	CERT_PATH             = LIMAN_PATH + "/certs"
)

var (
	location, _ = time.LoadLocation("Europe/Istanbul")

	GLOBAL_SCHEDULER = gocron.NewScheduler(location)
)
