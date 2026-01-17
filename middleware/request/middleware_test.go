package request

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

type TestRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
}

type TestURIRequest struct {
	ID string `uri:"id" binding:"required"`
}

func TestBindAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		request        interface{}
		expectedStatus int
		shouldAbort    bool
	}{
		{
			name: "valid request",
			request: TestRequest{
				Email:    "test@example.com",
				Password: "password123",
			},
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name: "invalid email",
			request: TestRequest{
				Email:    "invalid-email",
				Password: "password123",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			shouldAbort:    true,
		},
		{
			name: "password too short",
			request: TestRequest{
				Email:    "test@example.com",
				Password: "short",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			shouldAbort:    true,
		},
		{
			name: "missing required fields",
			request: TestRequest{
				Email:    "",
				Password: "",
			},
			expectedStatus: http.StatusUnprocessableEntity,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			body, _ := json.Marshal(tt.request)
			c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
			c.Request.Header.Set("Content-Type", "application/json")

			middleware := BindAndValidate[TestRequest]()
			middleware(c)

			if tt.shouldAbort && !c.IsAborted() {
				t.Error("expected context to be aborted")
			}

			if !tt.shouldAbort && c.IsAborted() {
				t.Error("expected context not to be aborted")
			}

			if tt.shouldAbort && w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestBindAndValidateWithInvalidJSON(t *testing.T) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	c.Request, _ = http.NewRequest(http.MethodPost, "/test", bytes.NewBufferString("invalid json"))
	c.Request.Header.Set("Content-Type", "application/json")

	middleware := BindAndValidate[TestRequest]()
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected context to be aborted")
	}

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestGetRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		expectedReq := &TestRequest{
			Email:    "test@example.com",
			Password: "password123",
		}

		c.Set(requestKeyCtx, expectedReq)

		req := GetRequest[TestRequest](c)

		if req == nil {
			t.Fatal("expected request to be returned")
		}

		if req.Email != expectedReq.Email {
			t.Errorf("expected email %q, got %q", expectedReq.Email, req.Email)
		}

		if req.Password != expectedReq.Password {
			t.Errorf("expected password %q, got %q", expectedReq.Password, req.Password)
		}
	})

	t.Run("request not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := GetRequest[TestRequest](c)

		if req != nil {
			t.Error("expected nil request")
		}

		if !c.IsAborted() {
			t.Error("expected context to be aborted")
		}

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})

	t.Run("invalid type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set wrong type
		c.Set(requestKeyCtx, "wrong type")

		req := GetRequest[TestRequest](c)

		if req != nil {
			t.Error("expected nil request")
		}

		if !c.IsAborted() {
			t.Error("expected context to be aborted")
		}

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestBindAndValidateURI(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name           string
		uri            string
		expectedStatus int
		shouldAbort    bool
	}{
		{
			name:           "valid URI parameter",
			uri:            "123",
			expectedStatus: http.StatusOK,
			shouldAbort:    false,
		},
		{
			name:           "empty URI parameter",
			uri:            "",
			expectedStatus: http.StatusNotFound,
			shouldAbort:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			c.Request, _ = http.NewRequest(http.MethodGet, "/test/"+tt.uri, nil)
			c.Params = []gin.Param{
				{Key: "id", Value: tt.uri},
			}

			middleware := BindAndValidateURI[TestURIRequest]()
			middleware(c)

			if tt.shouldAbort && !c.IsAborted() {
				t.Error("expected context to be aborted")
			}

			if !tt.shouldAbort && c.IsAborted() {
				t.Error("expected context not to be aborted")
			}

			if tt.shouldAbort && w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetUriRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("valid URI request", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		expectedReq := &TestURIRequest{
			ID: "123",
		}

		c.Set(uriRequestKeyCtx, expectedReq)

		req := GetUriRequest[TestURIRequest](c)

		if req == nil {
			t.Fatal("expected request to be returned")
		}

		if req.ID != expectedReq.ID {
			t.Errorf("expected ID %q, got %q", expectedReq.ID, req.ID)
		}
	})

	t.Run("URI request not found", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		req := GetUriRequest[TestURIRequest](c)

		if req != nil {
			t.Error("expected nil request")
		}

		if !c.IsAborted() {
			t.Error("expected context to be aborted")
		}

		if w.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, w.Code)
		}
	})

	t.Run("invalid URI type", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		// Set wrong type
		c.Set(uriRequestKeyCtx, "wrong type")

		req := GetUriRequest[TestURIRequest](c)

		if req != nil {
			t.Error("expected nil request")
		}

		if !c.IsAborted() {
			t.Error("expected context to be aborted")
		}

		if w.Code != http.StatusBadRequest {
			t.Errorf("expected status %d, got %d", http.StatusBadRequest, w.Code)
		}
	})
}

func TestBindAndValidateIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.POST("/test", BindAndValidate[TestRequest](), func(c *gin.Context) {
		req := GetRequest[TestRequest](c)
		if req == nil {
			t.Error("expected request to be available")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"email": req.Email,
		})
	})

	t.Run("full integration test", func(t *testing.T) {
		w := httptest.NewRecorder()
		body, _ := json.Marshal(TestRequest{
			Email:    "test@example.com",
			Password: "password123",
		})

		req, _ := http.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}

func TestBindAndValidateURIIntegration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	router.GET("/test/:id", BindAndValidateURI[TestURIRequest](), func(c *gin.Context) {
		req := GetUriRequest[TestURIRequest](c)
		if req == nil {
			t.Error("expected request to be available")
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"id": req.ID,
		})
	})

	t.Run("full URI integration test", func(t *testing.T) {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(http.MethodGet, "/test/123", nil)

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
	})
}
