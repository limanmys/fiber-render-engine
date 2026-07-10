package kubernetes

import (
	"encoding/base64"
	"fmt"
	"sync"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/clientcmd/api"
)

type Client struct {
	clientset *kubernetes.Clientset
	config    *rest.Config
}

var (
	clientCache = make(map[string]*Client)
	cacheMutex  = sync.RWMutex{}
)

func NewClient(kubeconfigBase64 string) (*Client, error) {
	// Check cache first
	cacheMutex.RLock()
	if client, exists := clientCache[kubeconfigBase64]; exists {
		cacheMutex.RUnlock()
		return client, nil
	}
	cacheMutex.RUnlock()

	// Create new client
	client, err := createClientFromBase64Config(kubeconfigBase64)
	if err != nil {
		return nil, err
	}

	// Cache the client
	cacheMutex.Lock()
	clientCache[kubeconfigBase64] = client
	cacheMutex.Unlock()

	return client, nil
}

func (kc *Client) GetClientset() *kubernetes.Clientset {
	return kc.clientset
}

func (kc *Client) GetConfig() *rest.Config {
	return kc.config
}

func createClientFromBase64Config(kubeconfigBase64 string) (*Client, error) {
	// Decode base64
	kubeconfigData, err := base64.StdEncoding.DecodeString(kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64 kubeconfig: %v", err)
	}

	// Parse kubeconfig directly from data without creating temp files
	config, err := clientcmd.Load(kubeconfigData)
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %v", err)
	}

	// Security: reject kubeconfigs that declare exec credential plugins.
	// The exec field allows running arbitrary commands on the host to obtain
	// credentials. Since kubeconfigs are user-supplied here, honoring exec
	// would permit remote command execution as the render engine user (root).
	if err := validateNoExecPlugins(config); err != nil {
		return nil, err
	}

	// Build config directly from the loaded config
	restConfig, err := buildConfigFromAPIConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to build rest config: %v", err)
	}

	// Create clientset
	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	return &Client{
		clientset: clientset,
		config:    restConfig,
	}, nil
}

func buildConfigFromAPIConfig(config *api.Config) (*rest.Config, error) {
	return clientcmd.NewDefaultClientConfig(*config, &clientcmd.ConfigOverrides{}).ClientConfig()
}

// validateNoExecPlugins ensures that no auth info in the kubeconfig relies on
// an exec credential plugin, which would execute arbitrary commands on the
// host when the client establishes a connection. This is unsafe for
// user-supplied kubeconfigs, so any such declaration is rejected outright.
func validateNoExecPlugins(config *api.Config) error {
	for name, authInfo := range config.AuthInfos {
		if authInfo == nil {
			continue
		}
		if authInfo.Exec != nil {
			return fmt.Errorf("kubeconfig auth info %q uses an exec credential plugin, which is not permitted for security reasons", name)
		}
	}
	return nil
}

// ClearCache clears the client cache (useful for testing or memory management)
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	clientCache = make(map[string]*Client)
}
