package middleware

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func TestIsValidationError(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		expectedValid bool
	}{
		{
			name:          "nil error",
			err:           nil,
			expectedValid: false,
		},
		{
			name:          "regular error",
			err:           errors.New("regular error"),
			expectedValid: false,
		},
		{
			name:          "validation error",
			err:           createValidationError(),
			expectedValid: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid, ve := IsValidationError(tt.err)

			if isValid != tt.expectedValid {
				t.Errorf("expected isValid %v, got %v", tt.expectedValid, isValid)
			}

			if tt.expectedValid && ve == nil {
				t.Error("expected validation errors to be returned")
			}

			if !tt.expectedValid && ve != nil {
				t.Error("expected validation errors to be nil")
			}
		})
	}
}

func TestValidatorErrorResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	// Create a validation error
	err := createValidationError()
	isValid, ve := IsValidationError(err)

	if !isValid {
		t.Fatal("expected validation error")
	}

	ValidatorErrorResponse(c, ve)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	if w.Body.Len() == 0 {
		t.Error("expected response body to be set")
	}

	// Check if response is aborted
	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}
}

// Helper function to create a validation error
func createValidationError() error {
	type TestStruct struct {
		Email string `validate:"required,email"`
		Age   int    `validate:"required,min=18"`
	}

	validate := validator.New()
	test := TestStruct{
		Email: "invalid-email",
		Age:   10,
	}

	return validate.Struct(test)
}

func TestValidationErrorResponseFormat(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	err := createValidationError()
	isValid, ve := IsValidationError(err)

	if !isValid {
		t.Fatal("expected validation error")
	}

	ValidatorErrorResponse(c, ve)

	var response ValidationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.Message != "The given data was invalid." {
		t.Fatalf("unexpected message: %s", response.Message)
	}

	if len(response.Errors) == 0 {
		t.Fatal("expected validation errors map to be non-empty")
	}

	// Verify content type
	contentType := w.Header().Get("Content-Type")
	if contentType != "application/json; charset=utf-8" {
		t.Errorf("expected content type 'application/json; charset=utf-8', got '%s'", contentType)
	}
}

func TestValidationErrorResponseWithMultipleErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type ComplexStruct struct {
		Email    string `validate:"required,email"`
		Age      int    `validate:"required,min=18,max=100"`
		Username string `validate:"required,min=3,max=20"`
		Password string `validate:"required,min=8"`
	}

	validate := validator.New()
	test := ComplexStruct{
		Email:    "",
		Age:      15,
		Username: "ab",
		Password: "123",
	}

	err := validate.Struct(test)
	isValid, ve := IsValidationError(err)

	if !isValid {
		t.Fatal("expected validation error")
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	ValidatorErrorResponse(c, ve)

	if w.Code != http.StatusUnprocessableEntity {
		t.Errorf("expected status %d, got %d", http.StatusUnprocessableEntity, w.Code)
	}

	var response ValidationErrorResponse
	if err := json.Unmarshal(w.Body.Bytes(), &response); err != nil {
		t.Fatalf("expected valid json response, got error: %v", err)
	}

	if response.Message != "The given data was invalid." {
		t.Fatalf("unexpected message: %s", response.Message)
	}

	if _, ok := response.Errors["email"]; !ok {
		t.Fatalf("expected email errors in response: %#v", response.Errors)
	}

	if _, ok := response.Errors["age"]; !ok {
		t.Fatalf("expected age errors in response: %#v", response.Errors)
	}

	if _, ok := response.Errors["username"]; !ok {
		t.Fatalf("expected username errors in response: %#v", response.Errors)
	}

	if _, ok := response.Errors["password"]; !ok {
		t.Fatalf("expected password errors in response: %#v", response.Errors)
	}
}
