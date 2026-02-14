package http

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type TransportServer struct {
	cfg    *cfg
	engine *gin.Engine
	server *http.Server
	mu     sync.RWMutex
}

func NewTransportServer(opts ...Option) *TransportServer {
	c := &cfg{
		host: "localhost",
		port: 8080,
		mode: MODE_PROD,
		authMiddleware: func(c *gin.Context) {
			c.Next()
		},
		permissionMiddleware: func(c *gin.Context) {
			c.Next()
		},
		corsMiddleware: corsMiddleware(),
	}

	for _, opt := range opts {
		opt(c)
	}

	switch c.mode {
	case MODE_DEV:
		gin.SetMode(gin.DebugMode)
	case MODE_TEST:
		gin.SetMode(gin.TestMode)
	case MODE_PROD:
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(c.corsMiddleware)

	return &TransportServer{
		cfg:    c,
		engine: engine,
	}
}

func (s *TransportServer) GetEngine() *gin.Engine {
	return s.engine
}

func (s *TransportServer) RegisterHandlers(handlers ...Handler) {
	s.engine.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusNotFound, gin.H{"message": "Not found"})
	})

	apiGroup := s.engine.Group("/api/v1")
	authProtectedGroup := apiGroup.Group("/")
	authProtectedGroup.Use(s.cfg.authMiddleware)
	if s.cfg.permissionMiddleware != nil {
		authProtectedGroup.Use(s.cfg.permissionMiddleware)
	}

	for _, handler := range handlers {
		for _, route := range handler.GetRoutes() {

			handlersChain := append([]gin.HandlerFunc{}, route.Middlewares...)
			handlersChain = append(handlersChain, route.Handler)

			switch route.Method {
			case http.MethodGet:
				if route.IsAuthProtected {
					authProtectedGroup.GET(route.Uri, handlersChain...)
				} else {
					apiGroup.GET(route.Uri, handlersChain...)
				}
			case http.MethodPost:
				if route.IsAuthProtected {
					authProtectedGroup.POST(route.Uri, handlersChain...)
				} else {
					apiGroup.POST(route.Uri, handlersChain...)
				}
			case http.MethodPut:
				if route.IsAuthProtected {
					authProtectedGroup.PUT(route.Uri, handlersChain...)
				} else {
					apiGroup.PUT(route.Uri, handlersChain...)
				}
			case http.MethodDelete:
				if route.IsAuthProtected {
					authProtectedGroup.DELETE(route.Uri, handlersChain...)
				} else {
					apiGroup.DELETE(route.Uri, handlersChain...)
				}
			}
		}
	}
}

func (s *TransportServer) Start() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.host, s.cfg.port)

	s.mu.Lock()
	s.server = &http.Server{
		Addr:              addr,
		Handler:           s.engine,
		ReadHeaderTimeout: 3 * time.Second,
	}
	srv := s.server
	s.mu.Unlock()

	return srv.ListenAndServe()
}

func (s *TransportServer) Stop(ctx context.Context) error {
	s.mu.RLock()
	srv := s.server
	s.mu.RUnlock()

	if srv == nil {
		return nil
	}

	return srv.Shutdown(ctx)
}

func corsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,PATCH,HEAD,OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin,Content-Type,Accept,Authorization,X-Requested-With")

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
