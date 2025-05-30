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

// ClearCache clears the client cache (useful for testing or memory management)
func ClearCache() {
	cacheMutex.Lock()
	defer cacheMutex.Unlock()
	clientCache = make(map[string]*Client)
}
