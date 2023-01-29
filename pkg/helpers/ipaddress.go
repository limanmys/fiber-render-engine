package helpers

import (
	"errors"
	"net"
)

// Resolve IP address from hostname
func ResolveIP(hostname string) (string, error) {
	for i := 0; i < 10; i++ {
		addr, err := net.LookupIP(hostname)
		if err == nil {
			return addr[0].String(), nil
		}
	}

	return "", errors.New("cannot resolve hostname")
}
