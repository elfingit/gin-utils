# Examples of usage

## Bind and validate request
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