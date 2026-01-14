package http

import "github.com/gin-gonic/gin"

type Route struct {
	Uri             string
	Method          string
	Handler         func(c *gin.Context)
	IsAuthProtected bool

	Middlewares []gin.HandlerFunc
}

type Handler interface {
	GetRoutes() []Route
}
