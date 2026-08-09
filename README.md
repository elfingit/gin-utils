# gin-utils

[![Tests](https://github.com/elfingit/gin-utils/actions/workflows/test.yml/badge.svg)](https://github.com/elfingit/gin-utils/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/elfingit/gin-utils)](https://goreportcard.com/report/github.com/elfingit/gin-utils)
[![codecov](https://codecov.io/github/elfingit/gin-utils/graph/badge.svg?token=T9F2RHFD5Q)](https://codecov.io/github/elfingit/gin-utils)

Utilities library for Gin Framework with support for request validation, response formatting, and route management.

## Features

- 🚀 Simple HTTP server setup with support for different modes (prod, dev, test)
- 🔒 Built-in middleware support for authentication and permission checking
- ✅ Automatic request validation and binding using generics
- 📦 Unified response format with envelope pattern support
- 🧪 Comprehensive test coverage (96.5%)

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
            Group:           "/api/v1/auth",
            Method:          http.MethodPost,
            IsAuthProtected: false,
            Uri:             "/login",
            Handler:         h.login,
            Middlewares: []gin.HandlerFunc{
                request.BindAndValidate[payload.LoginRequest](),
            },
        },
    }
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

### Route groups and auth

`Group` is the full route prefix counted from the root, so `Group: "/api/v1/users"` with `Uri: "/list"` is served at `/api/v1/users/list`. An empty `Group` registers the route at the root.

Routes are collected into gin router groups by their `Group` value, and `IsAuthProtected: true` puts a route into a separate protected group of the same prefix. The auth and permission middlewares are attached once to that group instead of being repeated for every handler, so protected and public routes can freely share the same prefix:

```go
func (h *UserHandler) GetRoutes() []pkghttp.Route {
    return []pkghttp.Route{
        {
            Group:           "/api/v1/users",
            Method:          http.MethodGet,
            Uri:             "/public",
            IsAuthProtected: false,
            Handler:         h.public,
        },
        {
            Group:           "/api/v1/users",
            Method:          http.MethodGet,
            Uri:             "/me",
            IsAuthProtected: true, // auth + permission middlewares run before the handler
            Handler:         h.me,
        },
    }
}
```

The resulting middleware order for a protected route is: CORS, auth, permission, route `Middlewares`, handler.

Supported methods are `GET`, `POST`, `PUT`, `PATCH`, `DELETE`, `HEAD` and `OPTIONS`. Method values are case-insensitive, and routes with an unsupported method are skipped during registration.

## CI/CD

The project uses GitHub Actions for continuous integration. On every pull request to master:
- Tests run on Go versions 1.25
- Code is checked with golangci-lint
- Test coverage is calculated and reported
- Merge is blocked if tests fail

See [Branch Protection Setup](.github/BRANCH_PROTECTION.md) for configuration details.

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

- **Main package**: 92.9%
- **middleware**: 100%
- **middleware/request**: 100%
- **middleware/response**: 100%
- **Total coverage**: 96.5%