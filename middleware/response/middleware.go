package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Envelope struct {
	Data  any            `json:"data,omitempty"`
	Error *ErrorResponse `json:"error,omitempty"`
	Meta  any            `json:"meta,omitempty"`
}

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func OK(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, Envelope{Data: data})
}

func WithMeta(c *gin.Context, data any, meta any) {
	c.JSON(http.StatusOK, Envelope{Data: data, Meta: meta})
}

func Fail(c *gin.Context, code int, message string) {

	var httpCode int

	if code == -1 {
		httpCode = http.StatusInternalServerError
	} else {
		httpCode = http.StatusBadRequest
	}

	c.JSON(httpCode, Envelope{Error: &ErrorResponse{Code: code, Message: message}})
}
