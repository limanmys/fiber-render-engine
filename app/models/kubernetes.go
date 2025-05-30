package models

// Request models for Kubernetes API
type KubeconfigRequest struct {
	Kubeconfig string `json:"kubeconfig" validate:"required,base64"`
}

type NamespaceRequest struct {
	Kubeconfig string `json:"kubeconfig" validate:"required,base64"`
}

type DeploymentRequest struct {
	Kubeconfig string `json:"kubeconfig" validate:"required,base64"`
	Namespace  string `json:"namespace" validate:"required"`
}

type DeploymentDetailsRequest struct {
	Kubeconfig string `json:"kubeconfig" validate:"required,base64"`
	Namespace  string `json:"namespace" validate:"required"`
	Deployment string `json:"deployment" validate:"required"`
}
type NamespaceInfo struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	CreationTime interface{}       `json:"creationTime"`
	Labels       map[string]string `json:"labels"`
	Annotations  map[string]string `json:"annotations"`
}

type DeploymentInfo struct {
	Name              string            `json:"name"`
	Namespace         string            `json:"namespace"`
	Replicas          int32             `json:"replicas"`
	ReadyReplicas     int32             `json:"readyReplicas"`
	AvailableReplicas int32             `json:"availableReplicas"`
	CreationTime      interface{}       `json:"creationTime"`
	Labels            map[string]string `json:"labels"`
	Annotations       map[string]string `json:"annotations"`
}

type PodInfo struct {
	Name              string                `json:"name"`
	Status            string                `json:"status"`
	CreationTime      interface{}           `json:"creationTime"`
	Node              string                `json:"node"`
	ContainerStatuses []ContainerStatusInfo `json:"containerStatuses"`
}

type ContainerStatusInfo struct {
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Started *bool  `json:"started"`
	Image   string `json:"image"`
}

type ReplicaSetInfo struct {
	Name           string      `json:"name"`
	Replicas       int32       `json:"replicas"`
	ReadyReplicas  int32       `json:"readyReplicas"`
	CreationTime   interface{} `json:"creationTime"`
	OwnerReference interface{} `json:"ownerReference"`
}

type ContainerInfo struct {
	Name    string              `json:"name"`
	Image   string              `json:"image"`
	Ports   []ContainerPortInfo `json:"ports"`
	Env     []ContainerEnvInfo  `json:"env"`
	Command []string            `json:"command"`
	Args    []string            `json:"args"`
}

type ContainerPortInfo struct {
	ContainerPort int32  `json:"containerPort"`
	Protocol      string `json:"protocol"`
	Name          string `json:"name"`
}

type ContainerEnvInfo struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type DeploymentDetails struct {
	Metadata    interface{}      `json:"metadata"`
	Spec        interface{}      `json:"spec"`
	Status      interface{}      `json:"status"`
	Pods        []PodInfo        `json:"pods"`
	ReplicaSets []ReplicaSetInfo `json:"replicaSets"`
}
