package local_access

import (
	"net"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/limanmys/render-engine/pkg/logger"
)

// Config defines the configuration for local access middleware
type Config struct {
	// AllowedNetworks defines the networks that are allowed to access the application
	// Default: []string{"127.0.0.1/32", "::1/128", "10.0.0.0/8", "172.16.0.0/12", "192.168.0.0/16"}
	AllowedNetworks []string

	// DenyMessage is the message returned when access is denied
	// Default: "Access denied: not from local network"
	DenyMessage string

	// StatusCode is the HTTP status code returned when access is denied
	// Default: 403 (Forbidden)
	StatusCode int

	// Next defines a function to skip this middleware when returned true.
	// Optional. Default: nil
	Next func(c *fiber.Ctx) bool
}

// ConfigDefault is the default config
var ConfigDefault = Config{
	AllowedNetworks: []string{
		"127.0.0.1/32",   // IPv4 localhost
		"::1/128",        // IPv6 localhost
		"10.0.0.0/8",     // Private network (Class A)
		"172.16.0.0/12",  // Private network (Class B)
		"192.168.0.0/16", // Private network (Class C)
	},
	DenyMessage: "Access denied: not from local network",
	StatusCode:  fiber.StatusForbidden,
	Next:        nil,
}

// Helper function to merge user config with default config
func configDefault(config ...Config) Config {
	if len(config) < 1 {
		return ConfigDefault
	}

	cfg := config[0]

	if cfg.AllowedNetworks == nil {
		cfg.AllowedNetworks = ConfigDefault.AllowedNetworks
	}

	if cfg.DenyMessage == "" {
		cfg.DenyMessage = ConfigDefault.DenyMessage
	}

	if cfg.StatusCode == 0 {
		cfg.StatusCode = ConfigDefault.StatusCode
	}

	return cfg
}

// Creates a new local access middleware handler.
// This middleware restricts access to certain routes based on the client's IP address.
// It checks if the client IP is within the allowed networks (localhost and private networks by default).
// Example usage:
// app.Use(local_access.New())
//
// Custom configuration:
//
//	app.Use(local_access.New(local_access.Config{
//	    AllowedNetworks: []string{"192.168.1.0/24", "10.0.0.0/8"},
//	    DenyMessage: "Access denied from your network",
//	    StatusCode: 403,
//	}))
func New(config ...Config) fiber.Handler {
	cfg := configDefault(config...)

	// Parse allowed networks into CIDR blocks
	var allowedCIDRs []*net.IPNet
	for _, network := range cfg.AllowedNetworks {
		// Handle single IP addresses by adding appropriate CIDR suffix
		if !strings.Contains(network, "/") {
			if strings.Contains(network, ":") {
				// IPv6 address
				network += "/128"
			} else {
				// IPv4 address
				network += "/32"
			}
		}

		_, cidr, err := net.ParseCIDR(network)
		if err != nil {
			logger.Sugar().Errorf("Invalid network CIDR '%s': %v", network, err)
			continue
		}
		allowedCIDRs = append(allowedCIDRs, cidr)
	}

	return func(c *fiber.Ctx) error {
		// Skip middleware if Next returns true
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}

		// Get client IP address
		clientIP := c.IP()
		if clientIP == "" {
			logger.Sugar().Warn("Unable to determine client IP address")
			return logger.FiberError(cfg.StatusCode, cfg.DenyMessage)
		}

		// Parse client IP
		ip := net.ParseIP(clientIP)
		if ip == nil {
			logger.Sugar().Warnf("Invalid client IP address: %s", clientIP)
			return logger.FiberError(cfg.StatusCode, cfg.DenyMessage)
		}

		// Check if client IP is in any of the allowed networks
		for _, cidr := range allowedCIDRs {
			if cidr.Contains(ip) {
				// IP is allowed, continue to next middleware
				return c.Next()
			}
		}

		// Log the denied access attempt
		logger.Sugar().Infof("Access denied for IP: %s", clientIP)

		// IP is not in allowed networks, deny access
		return logger.FiberError(cfg.StatusCode, cfg.DenyMessage)
	}
}
