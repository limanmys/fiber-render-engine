package kubernetes

import (
	"context"
	"fmt"

	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/pkg/kubernetes"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (k *Service) GetNamespaces(ctx context.Context, kubeconfigBase64 string) ([]models.NamespaceInfo, error) {
	kubeClient, err := kubernetes.NewClient(kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	namespaces, err := kubeClient.GetClientset().CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list namespaces: %v", err)
	}

	var namespaceList []models.NamespaceInfo
	for _, ns := range namespaces.Items {
		namespaceList = append(namespaceList, models.NamespaceInfo{
			Name:         ns.Name,
			Status:       string(ns.Status.Phase),
			CreationTime: ns.CreationTimestamp.Time,
			Labels:       ns.Labels,
			Annotations:  ns.Annotations,
		})
	}

	return namespaceList, nil
}

func (k *Service) GetDeployments(ctx context.Context, kubeconfigBase64, namespace string) ([]models.DeploymentInfo, error) {
	kubeClient, err := kubernetes.NewClient(kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	deployments, err := kubeClient.GetClientset().AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to list deployments: %v", err)
	}

	var deploymentList []models.DeploymentInfo
	for _, deployment := range deployments.Items {
		replicas := int32(0)
		if deployment.Spec.Replicas != nil {
			replicas = *deployment.Spec.Replicas
		}

		deploymentList = append(deploymentList, models.DeploymentInfo{
			Name:              deployment.Name,
			Namespace:         deployment.Namespace,
			Replicas:          replicas,
			ReadyReplicas:     deployment.Status.ReadyReplicas,
			AvailableReplicas: deployment.Status.AvailableReplicas,
			CreationTime:      deployment.CreationTimestamp.Time,
			Labels:            deployment.Labels,
			Annotations:       deployment.Annotations,
		})
	}

	return deploymentList, nil
}

func (k *Service) GetDeploymentDetails(ctx context.Context, kubeconfigBase64, namespace, deploymentName string) (*models.DeploymentDetails, error) {
	kubeClient, err := kubernetes.NewClient(kubeconfigBase64)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %v", err)
	}

	clientset := kubeClient.GetClientset()

	// Get deployment
	deployment, err := clientset.AppsV1().Deployments(namespace).Get(ctx, deploymentName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %v", err)
	}

	// Get pods related to this deployment
	pods, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get pods: %v", err)
	}

	var podList []models.PodInfo
	for _, pod := range pods.Items {
		var containerStatuses []models.ContainerStatusInfo
		for _, containerStatus := range pod.Status.ContainerStatuses {
			containerStatuses = append(containerStatuses, models.ContainerStatusInfo{
				Name:    containerStatus.Name,
				Ready:   containerStatus.Ready,
				Started: containerStatus.Started,
				Image:   containerStatus.Image,
			})
		}

		podList = append(podList, models.PodInfo{
			Name:              pod.Name,
			Status:            string(pod.Status.Phase),
			CreationTime:      pod.CreationTimestamp.Time,
			Node:              pod.Spec.NodeName,
			ContainerStatuses: containerStatuses,
		})
	}

	// Get replica sets related to this deployment
	replicaSets, err := clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{
		LabelSelector: metav1.FormatLabelSelector(deployment.Spec.Selector),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get replica sets: %v", err)
	}

	var replicaSetList []models.ReplicaSetInfo
	for _, rs := range replicaSets.Items {
		replicas := int32(0)
		if rs.Spec.Replicas != nil {
			replicas = *rs.Spec.Replicas
		}

		replicaSetList = append(replicaSetList, models.ReplicaSetInfo{
			Name:           rs.Name,
			Replicas:       replicas,
			ReadyReplicas:  rs.Status.ReadyReplicas,
			CreationTime:   rs.CreationTimestamp.Time,
			OwnerReference: rs.OwnerReferences,
		})
	}

	// Prepare container details
	var containers []models.ContainerInfo
	for _, container := range deployment.Spec.Template.Spec.Containers {
		var ports []models.ContainerPortInfo
		for _, port := range container.Ports {
			ports = append(ports, models.ContainerPortInfo{
				ContainerPort: port.ContainerPort,
				Protocol:      string(port.Protocol),
				Name:          port.Name,
			})
		}

		var env []models.ContainerEnvInfo
		for _, envVar := range container.Env {
			env = append(env, models.ContainerEnvInfo{
				Name:  envVar.Name,
				Value: envVar.Value,
			})
		}

		containers = append(containers, models.ContainerInfo{
			Name:    container.Name,
			Image:   container.Image,
			Ports:   ports,
			Env:     env,
			Command: container.Command,
			Args:    container.Args,
		})
	}

	deploymentDetails := &models.DeploymentDetails{
		Metadata: map[string]interface{}{
			"name":              deployment.Name,
			"namespace":         deployment.Namespace,
			"creationTimestamp": deployment.CreationTimestamp.Time,
			"labels":            deployment.Labels,
			"annotations":       deployment.Annotations,
			"uid":               deployment.UID,
		},
		Spec: map[string]interface{}{
			"replicas": func() int32 {
				if deployment.Spec.Replicas != nil {
					return *deployment.Spec.Replicas
				}
				return 0
			}(),
			"strategy": deployment.Spec.Strategy,
			"selector": deployment.Spec.Selector,
			"template": map[string]interface{}{
				"metadata": deployment.Spec.Template.ObjectMeta,
				"spec": map[string]interface{}{
					"containers": containers,
				},
			},
		},
		Status: map[string]interface{}{
			"replicas":            deployment.Status.Replicas,
			"readyReplicas":       deployment.Status.ReadyReplicas,
			"availableReplicas":   deployment.Status.AvailableReplicas,
			"unavailableReplicas": deployment.Status.UnavailableReplicas,
			"observedGeneration":  deployment.Status.ObservedGeneration,
			"conditions":          deployment.Status.Conditions,
		},
		Pods:        podList,
		ReplicaSets: replicaSetList,
	}

	return deploymentDetails, nil
}
