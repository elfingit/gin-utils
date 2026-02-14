package http

import (
	"testing"

	"github.com/gin-gonic/gin"
)

func TestWithHost(t *testing.T) {
	tests := []struct {
		name string
		host string
	}{
		{
			name: "set localhost",
			host: "localhost",
		},
		{
			name: "set custom host",
			host: "0.0.0.0",
		},
		{
			name: "set empty host",
			host: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cfg{}
			opt := WithHost(tt.host)
			opt(c)

			if c.host != tt.host {
				t.Errorf("expected host %q, got %q", tt.host, c.host)
			}
		})
	}
}

func TestWithPort(t *testing.T) {
	tests := []struct {
		name string
		port uint
	}{
		{
			name: "set default port",
			port: 8080,
		},
		{
			name: "set custom port",
			port: 3000,
		},
		{
			name: "set zero port",
			port: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cfg{}
			opt := WithPort(tt.port)
			opt(c)

			if c.port != tt.port {
				t.Errorf("expected port %d, got %d", tt.port, c.port)
			}
		})
	}
}

func TestWithMode(t *testing.T) {
	tests := []struct {
		name string
		mode string
	}{
		{
			name: "set production mode",
			mode: MODE_PROD,
		},
		{
			name: "set development mode",
			mode: MODE_DEV,
		},
		{
			name: "set test mode",
			mode: MODE_TEST,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &cfg{}
			opt := WithMode(tt.mode)
			opt(c)

			if c.mode != tt.mode {
				t.Errorf("expected mode %q, got %q", tt.mode, c.mode)
			}
		})
	}
}

func TestWithAuthMiddleware(t *testing.T) {
	called := false
	middleware := func(c *gin.Context) {
		called = true
		c.Next()
	}

	c := &cfg{}
	opt := WithAuthMiddleware(middleware)
	opt(c)

	if c.authMiddleware == nil {
		t.Error("expected authMiddleware to be set")
	}

	// Test middleware execution
	ctx := &gin.Context{}
	c.authMiddleware(ctx)

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestWithPermissionMiddleware(t *testing.T) {
	called := false
	middleware := func(c *gin.Context) {
		called = true
		c.Next()
	}

	c := &cfg{}
	opt := WithPermissionMiddleware(middleware)
	opt(c)

	if c.permissionMiddleware == nil {
		t.Error("expected permissionMiddleware to be set")
	}

	// Test middleware execution
	ctx := &gin.Context{}
	c.permissionMiddleware(ctx)

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestWithCorsMiddleware(t *testing.T) {
	called := false
	middleware := func(c *gin.Context) {
		called = true
		c.Next()
	}

	c := &cfg{}
	opt := WithCorsMiddleware(middleware)
	opt(c)

	if c.corsMiddleware == nil {
		t.Error("expected corsMiddleware to be set")
	}

	ctx := &gin.Context{}
	c.corsMiddleware(ctx)

	if !called {
		t.Error("expected middleware to be called")
	}
}

func TestMultipleOptions(t *testing.T) {
	expectedHost := "0.0.0.0"
	expectedPort := uint(9000)
	expectedMode := MODE_DEV

	c := &cfg{}
	opts := []Option{
		WithHost(expectedHost),
		WithPort(expectedPort),
		WithMode(expectedMode),
	}

	for _, opt := range opts {
		opt(c)
	}

	if c.host != expectedHost {
		t.Errorf("expected host %q, got %q", expectedHost, c.host)
	}
	if c.port != expectedPort {
		t.Errorf("expected port %d, got %d", expectedPort, c.port)
	}
	if c.mode != expectedMode {
		t.Errorf("expected mode %q, got %q", expectedMode, c.mode)
	}
}
