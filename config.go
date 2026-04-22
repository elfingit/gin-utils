package http

import (
	"context"
	"net"

	"github.com/gin-gonic/gin"
)

const (
	MODE_PROD = "prod"
	MODE_DEV  = "dev"
	MODE_TEST = "test"
)

type cfg struct {
	host                 string
	port                 uint
	mode                 string
	authMiddleware       func(c *gin.Context)
	permissionMiddleware func(c *gin.Context)
	corsMiddleware       func(c *gin.Context)
	baseContext          func(listener net.Listener) context.Context
}

type Option func(*cfg)

func WithHost(host string) Option {
	return func(c *cfg) {
		c.host = host
	}
}

func WithPort(port uint) Option {
	return func(c *cfg) {
		c.port = port
	}
}

func WithMode(mode string) Option {
	return func(c *cfg) {
		c.mode = mode
	}
}

func WithAuthMiddleware(middleware func(c *gin.Context)) Option {
	return func(c *cfg) {
		c.authMiddleware = middleware
	}
}

func WithPermissionMiddleware(middleware func(c *gin.Context)) Option {
	return func(c *cfg) {
		c.permissionMiddleware = middleware
	}
}

func WithCorsMiddleware(middleware func(c *gin.Context)) Option {
	return func(c *cfg) {
		c.corsMiddleware = middleware
	}
}

func WithBaseContext(baseContext func(listener net.Listener) context.Context) Option {
	return func(c *cfg) {
		c.baseContext = baseContext
	}
}
