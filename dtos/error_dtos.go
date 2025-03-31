package dtos

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewErrorResponse(errorMessage string) *ErrorResponse {
	return &ErrorResponse{Error: errorMessage}
}
