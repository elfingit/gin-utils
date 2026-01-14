package middleware

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}
