package middleware

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/go-playground/validator/v10"
)

func IsValidationError(err error) (bool, *validator.ValidationErrors) {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		return true, &ve
	}

	return false, nil
}

func ValidatorErrorResponse(c *gin.Context, err *validator.ValidationErrors) {
	vErr := *err
	out := make([]ValidationErrorResponse, 0, len(vErr))
	for _, fe := range vErr {
		out = append(out, ValidationErrorResponse{
			Field:   fe.Field(),
			Message: fe.Tag(),
		})
	}

	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, out)
}
