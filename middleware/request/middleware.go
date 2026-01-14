package request

import (
	"net/http"

	"github.com/elfingit/gin-utils/middleware"
	"github.com/gin-gonic/gin"
)

type requestKey struct{}
type uriRequestKey struct{}

var requestKeyCtx = requestKey{}
var uriRequestKeyCtx = uriRequestKey{}

func BindAndValidate[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T

		if err := c.ShouldBind(&req); err != nil {
			if ok, ve := middleware.IsValidationError(err); ok {
				middleware.ValidatorErrorResponse(c, ve)

				return
			}

			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"message": "Invalid request data",
				},
			})
			return
		}

		c.Set(requestKeyCtx, &req)
		c.Next()
	}
}

func BindAndValidateURI[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {
		var req T

		if err := c.ShouldBindUri(&req); err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"error": "Not found",
			})

			return
		}

		c.Set(uriRequestKeyCtx, &req)
		c.Next()
	}
}

func GetRequest[T any](c *gin.Context) *T {
	v, ok := c.Get(requestKeyCtx)
	if !ok {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
			"error": "Request not found",
		})
		return nil
	}

	if req, ok := v.(*T); ok {
		return req
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error": "Invalid type of request",
	})

	return nil
}

func GetUriRequest[T any](c *gin.Context) *T {
	v, ok := c.Get(uriRequestKeyCtx)
	if !ok {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"error": "Request not found",
		})
		return nil
	}

	if req, ok := v.(*T); ok {
		return req
	}

	c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
		"error": "Invalid type of request",
	})

	return nil
}
