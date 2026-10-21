package bridge

import (
	"bytes"
	"net"
	"strconv"
	"sync"
	"time"

	"github.com/limanmys/render-engine/app/models"
	"golang.org/x/crypto/ssh"
)

const rawSshHostKeyTTL = time.Hour

type rawSshHostKey struct {
	key       ssh.PublicKey
	expiresAt time.Time
}

// Raw connections trust the first key per endpoint for a fixed hour. This
// process-local store is separate from the managed servers' database policy.
type rawSshHostKeyCache struct {
	sync.Mutex
	keys        map[string]rawSshHostKey
	nextCleanup time.Time
}

var rawSshHostKeys = &rawSshHostKeyCache{}

func (c *rawSshHostKeyCache) cleanup(now time.Time) {
	if now.Before(c.nextCleanup) {
		return
	}
	for endpoint, entry := range c.keys {
		if !now.Before(entry.expiresAt) {
			delete(c.keys, endpoint)
		}
	}
	c.nextCleanup = now.Add(time.Minute)
}

func (c *rawSshHostKeyCache) trustedKeys(host string, port int, now time.Time) []models.SshHostKey {
	c.Lock()
	defer c.Unlock()
	c.cleanup(now)
	entry, ok := c.keys[net.JoinHostPort(normalizeSshHost(host), strconv.Itoa(port))]
	if !ok || !now.Before(entry.expiresAt) {
		return nil
	}
	return []models.SshHostKey{{PublicKey: string(ssh.MarshalAuthorizedKey(entry.key))}}
}

func (c *rawSshHostKeyCache) check(host string, port int, presented ssh.PublicKey, now time.Time) error {
	c.Lock()
	defer c.Unlock()
	c.cleanup(now)
	host = normalizeSshHost(host)
	endpoint := net.JoinHostPort(host, strconv.Itoa(port))
	entry, ok := c.keys[endpoint]
	if !ok || !now.Before(entry.expiresAt) {
		if c.keys == nil {
			c.keys = make(map[string]rawSshHostKey)
		}
		c.keys[endpoint] = rawSshHostKey{key: presented, expiresAt: now.Add(rawSshHostKeyTTL)}
		return nil
	}
	if bytes.Equal(entry.key.Marshal(), presented.Marshal()) {
		return nil
	}
	return &HostKeyError{Kind: hostKeyMismatch, Host: host, Port: port, Fingerprint: ssh.FingerprintSHA256(presented)}
}

func dialRawSsh(host, port string, config *ssh.ClientConfig) (*ssh.Client, error) {
	parsedPort, err := parseSshPort(port)
	if err != nil {
		return nil, err
	}
	host = normalizeSshHost(host)
	clientConfig := *config
	if keys := rawSshHostKeys.trustedKeys(host, parsedPort, time.Now()); len(keys) > 0 {
		algorithms, err := trustedHostKeyAlgorithms(keys)
		if err != nil {
			return nil, err
		}
		clientConfig.HostKeyAlgorithms = intersectHostKeyAlgorithms(algorithms, config.HostKeyAlgorithms)
	}
	clientConfig.HostKeyCallback = func(_ string, _ net.Addr, presented ssh.PublicKey) error {
		return rawSshHostKeys.check(host, parsedPort, presented, time.Now())
	}
	return dialSshWithConfig(host, strconv.Itoa(parsedPort), &clientConfig)
}
