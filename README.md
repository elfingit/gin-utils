# gin-utils

Utilities library for Gin Framework with support for request validation, response formatting, and route management.

## Features

- 🚀 Simple HTTP server setup with support for different modes (prod, dev, test)
- 🔒 Built-in middleware support for authentication and permission checking
- ✅ Automatic request validation and binding using generics
- 📦 Unified response format with envelope pattern support
- 🧪 Comprehensive test coverage (95.9%)

## Examples of usage

### Bind and validate request
```go
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8,max=64"`
}

...

func (h *AuthHandler) GetRoutes() []pkghttp.Route {
    return []pkghttp.Route{
        {
            Method:          http.MethodPost,
            IsAuthProtected: false,
            Uri:             "/login",
            Handler:         h.login,
            Middlewares: []gin.HandlerFunc{
            request.BindAndValidate[payload.LoginRequest](),
        },
    },
}

...
req := request.GetRequest[payload.LoginRequest](c)
if req == nil {
    c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
    "message": "empty request",
    })

    return
}
...	
	
```

## Testing

The project has comprehensive test coverage. To run tests use:

```bash
# Run all tests
go test -v ./...

# Run tests with coverage
go test -coverprofile=coverage.out ./...

# View detailed coverage report
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Test Coverage

- **Main package**: 91.7%
- **middleware**: 100%
- **middleware/request**: 100%
- **middleware/response**: 100%
- **Total coverage**: 95.9%