package http

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

type mockHandler struct {
	routes []Route
}

func (m *mockHandler) GetRoutes() []Route {
	return m.routes
}

func TestNewTransportServer(t *testing.T) {
	tests := []struct {
		name         string
		opts         []Option
		expectedHost string
		expectedPort uint
		expectedMode string
	}{
		{
			name:         "default configuration",
			opts:         []Option{},
			expectedHost: "localhost",
			expectedPort: 8080,
			expectedMode: MODE_PROD,
		},
		{
			name: "custom host and port",
			opts: []Option{
				WithHost("0.0.0.0"),
				WithPort(3000),
			},
			expectedHost: "0.0.0.0",
			expectedPort: 3000,
			expectedMode: MODE_PROD,
		},
		{
			name: "test mode",
			opts: []Option{
				WithMode(MODE_TEST),
			},
			expectedHost: "localhost",
			expectedPort: 8080,
			expectedMode: MODE_TEST,
		},
		{
			name: "dev mode",
			opts: []Option{
				WithMode(MODE_DEV),
			},
			expectedHost: "localhost",
			expectedPort: 8080,
			expectedMode: MODE_DEV,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewTransportServer(tt.opts...)

			if server == nil {
				t.Fatal("expected server to be created")
			}

			if server.cfg.host != tt.expectedHost {
				t.Errorf("expected host %q, got %q", tt.expectedHost, server.cfg.host)
			}

			if server.cfg.port != tt.expectedPort {
				t.Errorf("expected port %d, got %d", tt.expectedPort, server.cfg.port)
			}

			if server.cfg.mode != tt.expectedMode {
				t.Errorf("expected mode %q, got %q", tt.expectedMode, server.cfg.mode)
			}

			if server.engine == nil {
				t.Error("expected engine to be initialized")
			}
		})
	}
}

func TestTransportServerRegisterHandlers(t *testing.T) {
	tests := []struct {
		name            string
		routes          []Route
		expectedMethod  string
		expectedPath    string
		expectedStatus  int
		isAuthProtected bool
	}{
		{
			name: "register GET route",
			routes: []Route{
				{
					Group:           "/api/v1",
					Uri:             "/test",
					Method:          http.MethodGet,
					IsAuthProtected: false,
					Handler: func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"message": "success"})
					},
				},
			},
			expectedMethod:  http.MethodGet,
			expectedPath:    "/api/v1/test",
			expectedStatus:  http.StatusOK,
			isAuthProtected: false,
		},
		{
			name: "register POST route",
			routes: []Route{
				{
					Group:           "/api/v1",
					Uri:             "/create",
					Method:          http.MethodPost,
					IsAuthProtected: false,
					Handler: func(c *gin.Context) {
						c.JSON(http.StatusCreated, gin.H{"message": "created"})
					},
				},
			},
			expectedMethod:  http.MethodPost,
			expectedPath:    "/api/v1/create",
			expectedStatus:  http.StatusCreated,
			isAuthProtected: false,
		},
		{
			name: "register PUT route",
			routes: []Route{
				{
					Group:           "/api/v1",
					Uri:             "/update",
					Method:          http.MethodPut,
					IsAuthProtected: false,
					Handler: func(c *gin.Context) {
						c.JSON(http.StatusOK, gin.H{"message": "updated"})
					},
				},
			},
			expectedMethod:  http.MethodPut,
			expectedPath:    "/api/v1/update",
			expectedStatus:  http.StatusOK,
			isAuthProtected: false,
		},
		{
			name: "register DELETE route",
			routes: []Route{
				{
					Group:           "/api/v1",
					Uri:             "/delete",
					Method:          http.MethodDelete,
					IsAuthProtected: false,
					Handler: func(c *gin.Context) {
						c.JSON(http.StatusNoContent, gin.H{})
					},
				},
			},
			expectedMethod:  http.MethodDelete,
			expectedPath:    "/api/v1/delete",
			expectedStatus:  http.StatusNoContent,
			isAuthProtected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewTransportServer(WithMode(MODE_TEST))
			handler := &mockHandler{routes: tt.routes}

			server.RegisterHandlers(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.expectedMethod, tt.expectedPath, nil)
			server.engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestTransportServerAuthProtectedRoute(t *testing.T) {
	authCalled := false
	authMiddleware := func(c *gin.Context) {
		authCalled = true
		token := c.GetHeader("Authorization")
		if token != "valid-token" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}

	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithAuthMiddleware(authMiddleware),
	)

	handler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/protected",
				Method:          http.MethodGet,
				IsAuthProtected: true,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "protected data"})
				},
			},
		},
	}

	server.RegisterHandlers(handler)

	t.Run("unauthorized request", func(t *testing.T) {
		authCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		server.engine.ServeHTTP(w, req)

		if !authCalled {
			t.Error("expected auth middleware to be called")
		}

		if w.Code != http.StatusUnauthorized {
			t.Errorf("expected status %d, got %d", http.StatusUnauthorized, w.Code)
		}
	})

	t.Run("authorized request", func(t *testing.T) {
		authCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/protected", nil)
		req.Header.Set("Authorization", "valid-token")
		server.engine.ServeHTTP(w, req)

		if !authCalled {
			t.Error("expected auth middleware to be called")
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestTransportServerNotFoundRoute(t *testing.T) {
	server := NewTransportServer(WithMode(MODE_TEST))
	server.RegisterHandlers()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/nonexistent", nil)
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
	}
}

func TestTransportServerWithMiddlewares(t *testing.T) {
	middleware1Called := false
	middleware2Called := false

	middleware1 := func(c *gin.Context) {
		middleware1Called = true
		c.Next()
	}

	middleware2 := func(c *gin.Context) {
		middleware2Called = true
		c.Next()
	}

	server := NewTransportServer(WithMode(MODE_TEST))
	handler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/test",
				Method:          http.MethodGet,
				IsAuthProtected: false,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "success"})
				},
				Middlewares: []gin.HandlerFunc{middleware1, middleware2},
			},
		},
	}

	server.RegisterHandlers(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/test", nil)
	server.engine.ServeHTTP(w, req)

	if !middleware1Called {
		t.Error("expected middleware1 to be called")
	}

	if !middleware2Called {
		t.Error("expected middleware2 to be called")
	}

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTransportServerStartStop(t *testing.T) {
	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithHost("localhost"),
		WithPort(18080), // Fixed port for testing
	)

	server.RegisterHandlers()

	// Start server in goroutine
	errChan := make(chan error, 1)
	go func() {
		errChan <- server.Start()
	}()

	// Wait for server to be ready by attempting connections
	serverReady := false
	for i := 0; i < 50; i++ {
		time.Sleep(20 * time.Millisecond)
		server.mu.RLock()
		ready := server.server != nil
		server.mu.RUnlock()
		if ready {
			serverReady = true
			break
		}
	}

	if !serverReady {
		t.Fatal("server did not start in time")
	}

	// Stop server
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("expected no error on stop, got %v", err)
	}

	// Wait for Start to return
	select {
	case err := <-errChan:
		if err != nil && err != http.ErrServerClosed {
			t.Errorf("expected ErrServerClosed, got %v", err)
		}
	case <-time.After(2 * time.Second):
		t.Error("server did not stop in time")
	}
}

func TestTransportServerStopWithoutStart(t *testing.T) {
	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithHost("localhost"),
		WithPort(18081),
	)

	server.RegisterHandlers()

	// Try to stop server that was never started
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := server.Stop(ctx)
	if err != nil {
		t.Errorf("expected no error when stopping non-started server, got %v", err)
	}
}

func TestTransportServerGetEngine(t *testing.T) {
	server := NewTransportServer(WithMode(MODE_TEST))

	if server.GetEngine() == nil {
		t.Fatal("expected engine to be initialized")
	}

	if server.GetEngine() != server.engine {
		t.Error("expected GetEngine to return internal engine instance")
	}
}

func TestTransportServerGlobalCORSMiddleware(t *testing.T) {
	server := NewTransportServer(WithMode(MODE_TEST))

	t.Run("adds CORS headers to regular request", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/unknown", nil)
		server.engine.ServeHTTP(w, req)

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("expected Access-Control-Allow-Origin to be '*', got %q", got)
		}

		if got := w.Header().Get("Access-Control-Allow-Methods"); got == "" {
			t.Error("expected Access-Control-Allow-Methods to be set")
		}
	})

	t.Run("handles preflight OPTIONS globally", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodOptions, "/unknown", nil)
		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusNoContent {
			t.Errorf("expected status %d, got %d", http.StatusNoContent, w.Code)
		}

		if got := w.Header().Get("Access-Control-Allow-Origin"); got != "*" {
			t.Errorf("expected Access-Control-Allow-Origin to be '*', got %q", got)
		}
	})
}

func TestTransportServerCustomCORSMiddleware(t *testing.T) {
	customCORS := func(c *gin.Context) {
		c.Writer.Header().Set("X-Custom-CORS", "enabled")
		c.Next()
	}

	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithCorsMiddleware(customCORS),
	)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/unknown", nil)
	server.engine.ServeHTTP(w, req)

	if got := w.Header().Get("X-Custom-CORS"); got != "enabled" {
		t.Errorf("expected X-Custom-CORS to be 'enabled', got %q", got)
	}
}

func TestTransportServerAuthProtectedNonGetRoutes(t *testing.T) {
	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "auth protected POST route",
			method: http.MethodPost,
			path:   "/api/v1/protected-post",
		},
		{
			name:   "auth protected PUT route",
			method: http.MethodPut,
			path:   "/api/v1/protected-put",
		},
		{
			name:   "auth protected DELETE route",
			method: http.MethodDelete,
			path:   "/api/v1/protected-delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewTransportServer(WithMode(MODE_TEST))

			handler := &mockHandler{
				routes: []Route{
					{
						Group:           "/api/v1",
						Uri:             tt.path[len("/api/v1"):],
						Method:          tt.method,
						IsAuthProtected: true,
						Handler: func(c *gin.Context) {
							c.JSON(http.StatusOK, gin.H{"message": "ok"})
						},
					},
				},
			}

			server.RegisterHandlers(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.method, tt.path, nil)
			server.engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}
		})
	}
}

func TestTransportServerPermissionMiddleware(t *testing.T) {
	permissionCalled := false
	permissionMiddleware := func(c *gin.Context) {
		permissionCalled = true
		role := c.GetHeader("X-Role")
		if role != "admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			return
		}
		c.Next()
	}

	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithAuthMiddleware(func(c *gin.Context) {
			c.Next()
		}),
		WithPermissionMiddleware(permissionMiddleware),
	)

	handler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/admin",
				Method:          http.MethodGet,
				IsAuthProtected: true,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "admin data"})
				},
			},
		},
	}

	server.RegisterHandlers(handler)

	t.Run("forbidden request", func(t *testing.T) {
		permissionCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin", nil)
		req.Header.Set("X-Role", "user")
		server.engine.ServeHTTP(w, req)

		if !permissionCalled {
			t.Error("expected permission middleware to be called")
		}

		if w.Code != http.StatusForbidden {
			t.Errorf("expected status %d, got %d", http.StatusForbidden, w.Code)
		}
	})

	t.Run("allowed request", func(t *testing.T) {
		permissionCalled = false
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/admin", nil)
		req.Header.Set("X-Role", "admin")
		server.engine.ServeHTTP(w, req)

		if !permissionCalled {
			t.Error("expected permission middleware to be called")
		}

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestTransportServerDifferentGroups(t *testing.T) {
	server := NewTransportServer(WithMode(MODE_TEST))

	handler := &mockHandler{
		routes: []Route{
			{
				Group:  "/api/v1/users",
				Uri:    "/list",
				Method: http.MethodGet,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"version": "v1"})
				},
			},
			{
				Group:  "/api/v2/users",
				Uri:    "/list",
				Method: http.MethodGet,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"version": "v2"})
				},
			},
		},
	}

	server.RegisterHandlers(handler)

	tests := []struct {
		name         string
		path         string
		expectedBody string
	}{
		{
			name:         "v1 group",
			path:         "/api/v1/users/list",
			expectedBody: `{"version":"v1"}`,
		},
		{
			name:         "v2 group",
			path:         "/api/v2/users/list",
			expectedBody: `{"version":"v2"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, tt.path, nil)
			server.engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if w.Body.String() != tt.expectedBody {
				t.Errorf("expected body %q, got %q", tt.expectedBody, w.Body.String())
			}
		})
	}
}

func TestTransportServerMixedAuthInSameGroup(t *testing.T) {
	authCalls := 0
	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithAuthMiddleware(func(c *gin.Context) {
			authCalls++
			c.Next()
		}),
	)

	handler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/public",
				Method:          http.MethodGet,
				IsAuthProtected: false,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "public"})
				},
			},
			{
				Group:           "/api/v1",
				Uri:             "/private",
				Method:          http.MethodGet,
				IsAuthProtected: true,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "private"})
				},
			},
		},
	}

	server.RegisterHandlers(handler)

	t.Run("public route skips auth middleware", func(t *testing.T) {
		authCalls = 0
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/public", nil)
		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		if authCalls != 0 {
			t.Errorf("expected auth middleware not to be called, got %d calls", authCalls)
		}
	})

	t.Run("protected route runs auth middleware", func(t *testing.T) {
		authCalls = 0
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/api/v1/private", nil)
		server.engine.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}

		if authCalls != 1 {
			t.Errorf("expected auth middleware to be called once, got %d calls", authCalls)
		}
	})
}

func TestTransportServerReusesAuthProtectedGroup(t *testing.T) {
	authCalls := 0
	permissionCalls := 0

	server := NewTransportServer(
		WithMode(MODE_TEST),
		WithAuthMiddleware(func(c *gin.Context) {
			authCalls++
			c.Next()
		}),
		WithPermissionMiddleware(func(c *gin.Context) {
			permissionCalls++
			c.Next()
		}),
	)

	firstHandler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/first",
				Method:          http.MethodGet,
				IsAuthProtected: true,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "first"})
				},
			},
		},
	}

	secondHandler := &mockHandler{
		routes: []Route{
			{
				Group:           "/api/v1",
				Uri:             "/second",
				Method:          http.MethodGet,
				IsAuthProtected: true,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "second"})
				},
			},
		},
	}

	server.RegisterHandlers(firstHandler, secondHandler)

	for _, path := range []string{"/api/v1/first", "/api/v1/second"} {
		t.Run(path, func(t *testing.T) {
			authCalls = 0
			permissionCalls = 0

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(http.MethodGet, path, nil)
			server.engine.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			if authCalls != 1 {
				t.Errorf("expected auth middleware to be called once, got %d calls", authCalls)
			}

			if permissionCalls != 1 {
				t.Errorf("expected permission middleware to be called once, got %d calls", permissionCalls)
			}
		})
	}
}

func TestTransportServerEmptyGroup(t *testing.T) {
	server := NewTransportServer(WithMode(MODE_TEST))

	handler := &mockHandler{
		routes: []Route{
			{
				Uri:    "/login",
				Method: http.MethodPost,
				Handler: func(c *gin.Context) {
					c.JSON(http.StatusOK, gin.H{"message": "logged in"})
				},
			},
		},
	}

	server.RegisterHandlers(handler)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/login", nil)
	server.engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
	}
}

func TestTransportServerMethodHandling(t *testing.T) {
	tests := []struct {
		name           string
		routeMethod    string
		requestMethod  string
		expectedStatus int
	}{
		{
			name:           "PATCH route is registered",
			routeMethod:    http.MethodPatch,
			requestMethod:  http.MethodPatch,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "lowercase method is normalized",
			routeMethod:    "get",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "unsupported method is skipped",
			routeMethod:    "FOO",
			requestMethod:  http.MethodGet,
			expectedStatus: http.StatusNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewTransportServer(WithMode(MODE_TEST))

			handler := &mockHandler{
				routes: []Route{
					{
						Group:  "/api/v1",
						Uri:    "/resource",
						Method: tt.routeMethod,
						Handler: func(c *gin.Context) {
							c.JSON(http.StatusOK, gin.H{"message": "ok"})
						},
					},
				},
			}

			server.RegisterHandlers(handler)

			w := httptest.NewRecorder()
			req, _ := http.NewRequest(tt.requestMethod, "/api/v1/resource", nil)
			server.engine.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}
