package common

type APIError struct {
	StatusCode int
	Message    string
	InnerError error
}

func NewAPIError(statusCode int, message string, innerError error) *APIError {
	return &APIError{
		StatusCode: statusCode,
		Message:    message,
		InnerError: innerError,
	}
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) Unwrap() error {
	return e.InnerError
}
