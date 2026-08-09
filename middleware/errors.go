package middleware

import (
	"errors"
	"net/http"
	"strings"
	"unicode"

	"github.com/gin-gonic/gin"

	"github.com/go-playground/validator/v10"
)

const validationFailedMessage = "The given data was invalid."

func IsValidationError(err error) (bool, *validator.ValidationErrors) {
	var ve validator.ValidationErrors

	if errors.As(err, &ve) {
		return true, &ve
	}

	return false, nil
}

func ValidatorErrorResponse(c *gin.Context, err *validator.ValidationErrors) {
	vErr := *err
	out := ValidationErrorResponse{
		Message: validationFailedMessage,
		Errors:  make(map[string][]string, len(vErr)),
	}

	for _, fe := range vErr {
		fieldName := toLowerFirst(fe.Field())
		switch fe.Tag() {
		case "required":
			out.Errors[fieldName] = append(out.Errors[fieldName], "The "+fieldName+" field is required.")
		case "email":
			out.Errors[fieldName] = append(out.Errors[fieldName], "The "+fieldName+" field must be a valid email address.")
		default:
			out.Errors[fieldName] = append(out.Errors[fieldName], "The "+fieldName+" field is invalid.")
		}
	}

	c.AbortWithStatusJSON(http.StatusUnprocessableEntity, out)
}

func toLowerFirst(value string) string {
	if value == "" {
		return value
	}

	runes := []rune(value)
	runes[0] = unicode.ToLower(runes[0])
	return strings.TrimSpace(string(runes))
}
