package logger

/*
	This file represents the endpoints that must be logged
*/

var (
	// MINIMAL = 1
	MINIMAL []string = []string{}

	// EXT_LOG = 2
	EXT_LOG []string = []string{
		"/",
		"/externalAPI",
	}

	// EXT_DETAIL = 3
	EXT_DETAIL []string = []string{
		"/",
		"/externalAPI",
		"/command",
		"/outsideCommand",
		"/script",
		"/setExtensionDb",
		"/backgroundJob",
		"/openTunnel",
		"/getFile",
		"/putFile",
	}

	// ALL = 0
	ALL []string = []string{
		"/",
		"/command",
		"/outsideCommand",
		"/openTunnel",
		"/keepTunnelAlive",
		"/getFile",
		"/putFile",
		"/script",
		"/verify",
		"/setExtensionDb",
		"/sendLog",
		"/backgroundJob",
		"/externalAPI",
	}
)
