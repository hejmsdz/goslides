package common

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

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

func ReturnAPIError(c *gin.Context, statusCode int, message string, innerError error) {
	ReturnError(c, NewAPIError(statusCode, message, innerError))
}

func ReturnBadRequestError(c *gin.Context, innerError error) {
	ReturnError(c, NewAPIError(http.StatusBadRequest, "invalid request body", innerError))
}

func ReturnError(c *gin.Context, err error) {
	log.Printf("Aborting with error: %v", err)

	if apiErr, ok := err.(*APIError); ok {
		c.AbortWithStatusJSON(apiErr.StatusCode, gin.H{"error": apiErr.Message})
	} else {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
	}
}
