package response

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestOK(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name         string
		data         interface{}
		meta         interface{}
		expectedData interface{}
	}{
		{
			name:         "simple string data",
			data:         "test data",
			meta:         nil,
			expectedData: "test data",
		},
		{
			name: "map data",
			data: map[string]interface{}{
				"key": "value",
			},
			meta: nil,
			expectedData: map[string]interface{}{
				"key": "value",
			},
		},
		{
			name: "struct data",
			data: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "John",
				Age:  30,
			},
			meta: nil,
			expectedData: map[string]interface{}{
				"name": "John",
				"age":  float64(30),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			OK(c, tt.data, tt.meta)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response Envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Error != nil {
				t.Error("expected error to be nil")
			}

			if response.Meta != nil {
				t.Error("expected meta to be nil")
			}

			if response.Data == nil {
				t.Error("expected data to be set")
			}
		})
	}
}

func TestWithMeta(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		data interface{}
		meta interface{}
	}{
		{
			name: "with pagination meta",
			data: []string{"item1", "item2"},
			meta: map[string]interface{}{
				"page":  1,
				"total": 10,
			},
		},
		{
			name: "with custom meta",
			data: "data",
			meta: struct {
				Count int `json:"count"`
			}{
				Count: 5,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			WithMeta(c, tt.data, tt.meta)

			if w.Code != http.StatusOK {
				t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
			}

			var response Envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Error != nil {
				t.Error("expected error to be nil")
			}

			if response.Meta == nil {
				t.Error("expected meta to be set")
			}

			if response.Data == nil {
				t.Error("expected data to be set")
			}
		})
	}
}

func TestFail(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name               string
		code               int
		message            string
		expectedHTTPStatus int
	}{
		{
			name:               "internal server error",
			code:               -1,
			message:            "Internal server error",
			expectedHTTPStatus: http.StatusInternalServerError,
		},
		{
			name:               "bad request",
			code:               1001,
			message:            "Invalid input",
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "custom error code",
			code:               5000,
			message:            "Custom error",
			expectedHTTPStatus: http.StatusBadRequest,
		},
		{
			name:               "zero code",
			code:               0,
			message:            "Zero code error",
			expectedHTTPStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			Fail(c, tt.code, tt.message)

			if w.Code != tt.expectedHTTPStatus {
				t.Errorf("expected status %d, got %d", tt.expectedHTTPStatus, w.Code)
			}

			var response Envelope
			err := json.Unmarshal(w.Body.Bytes(), &response)
			if err != nil {
				t.Fatalf("failed to unmarshal response: %v", err)
			}

			if response.Data != nil {
				t.Error("expected data to be nil")
			}

			if response.Meta != nil {
				t.Error("expected meta to be nil")
			}

			if response.Error == nil {
				t.Fatal("expected error to be set")
			}

			if response.Error.Code != tt.code {
				t.Errorf("expected error code %d, got %d", tt.code, response.Error.Code)
			}

			if response.Error.Message != tt.message {
				t.Errorf("expected error message %q, got %q", tt.message, response.Error.Message)
			}
		})
	}
}

func TestEnvelopeStructure(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("OK response structure", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		testData := map[string]string{"key": "value"}
		OK(c, testData, nil)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if _, ok := response["data"]; !ok {
			t.Error("expected 'data' field in response")
		}

		if _, ok := response["error"]; ok {
			t.Error("expected 'error' field to be omitted")
		}

		if _, ok := response["meta"]; ok {
			t.Error("expected 'meta' field to be omitted")
		}
	})

	t.Run("WithMeta response structure", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		testData := map[string]string{"key": "value"}
		testMeta := map[string]int{"page": 1}
		WithMeta(c, testData, testMeta)

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if _, ok := response["data"]; !ok {
			t.Error("expected 'data' field in response")
		}

		if _, ok := response["meta"]; !ok {
			t.Error("expected 'meta' field in response")
		}

		if _, ok := response["error"]; ok {
			t.Error("expected 'error' field to be omitted")
		}
	})

	t.Run("Fail response structure", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		Fail(c, 1001, "Test error")

		var response map[string]interface{}
		err := json.Unmarshal(w.Body.Bytes(), &response)
		if err != nil {
			t.Fatalf("failed to unmarshal response: %v", err)
		}

		if _, ok := response["error"]; !ok {
			t.Error("expected 'error' field in response")
		}

		if _, ok := response["data"]; ok {
			t.Error("expected 'data' field to be omitted")
		}

		if _, ok := response["meta"]; ok {
			t.Error("expected 'meta' field to be omitted")
		}
	})
}

func TestContentType(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		fn   func(*gin.Context)
	}{
		{
			name: "OK",
			fn: func(c *gin.Context) {
				OK(c, "data", nil)
			},
		},
		{
			name: "WithMeta",
			fn: func(c *gin.Context) {
				WithMeta(c, "data", "meta")
			},
		},
		{
			name: "Fail",
			fn: func(c *gin.Context) {
				Fail(c, 1001, "error")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			tt.fn(c)

			contentType := w.Header().Get("Content-Type")
			if contentType != "application/json; charset=utf-8" {
				t.Errorf("expected content type 'application/json; charset=utf-8', got '%s'", contentType)
			}
		})
	}
}
