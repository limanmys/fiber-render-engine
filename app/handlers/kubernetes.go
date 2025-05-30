package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/app/models"
	"github.com/limanmys/render-engine/internal/kubernetes"
	"github.com/limanmys/render-engine/pkg/logger"
	"github.com/limanmys/render-engine/pkg/validator"
)

var kubernetesService = kubernetes.NewService()

// GetNamespaces lists all namespaces in the Kubernetes cluster
func GetNamespaces(c *fiber.Ctx) error {
	var req models.NamespaceRequest
	if err := c.BodyParser(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validator.ValidateStruct(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, err.Error())
	}

	namespaces, err := kubernetesService.GetNamespaces(c.Context(), req.Kubeconfig)
	if err != nil {
		return logger.FiberError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(namespaces)
}

// GetDeployments lists all deployments in a specific namespace
func GetDeployments(c *fiber.Ctx) error {
	var req models.DeploymentRequest
	if err := c.BodyParser(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validator.ValidateStruct(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, err.Error())
	}

	deployments, err := kubernetesService.GetDeployments(c.Context(), req.Kubeconfig, req.Namespace)
	if err != nil {
		return logger.FiberError(fiber.StatusInternalServerError, err.Error())
	}

	return c.JSON(deployments)
}

// GetDeploymentDetails gets detailed information about a specific deployment
func GetDeploymentDetails(c *fiber.Ctx) error {
	var req models.DeploymentDetailsRequest
	if err := c.BodyParser(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := validator.ValidateStruct(&req); err != nil {
		return logger.FiberError(fiber.StatusBadRequest, err.Error())
	}

	deploymentDetails, err := kubernetesService.GetDeploymentDetails(c.Context(), req.Kubeconfig, req.Namespace, req.Deployment)
	if err != nil {
		return logger.FiberError(fiber.StatusNotFound, err.Error())
	}

	return c.JSON(deploymentDetails)
}
