package errors

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

type ServiceError struct {
	Code    int
	Message string
	Fields  map[string]string 
	Err     error
}

func (e *ServiceError) Error() string {
	return e.Message
}

func NewValidationError(err error) *ServiceError {
    var ve validator.ValidationErrors
    if errors.As(err, &ve) {
        fields := make(map[string]string, len(ve))
        for _, fe := range ve {
            fields[fe.Field()] = validationMessage(fe)
        }

        return &ServiceError{
            Code:    http.StatusBadRequest,
            Message: "Validation failed",
            Fields:  fields,
        }
    }
    return &ServiceError{
        Code:    http.StatusBadRequest,
        Message: err.Error(),
    }
}

func validationMessage(fe validator.FieldError) string {
	switch fe.Tag() {
	case "required":
		return "This field is required"
	case "email":
		return "Must be a valid email address"
	case "min":
		n := fe.Param()
		if n == "1" {
			return "Must be at least 1 character"
		}
		return fmt.Sprintf("Must be at least %s characters", n)
	case "max":
		n := fe.Param()
		if n == "1" {
			return "Must be at most 1 character"
		}
		return fmt.Sprintf("Must be at most %s characters", n)
	case "oneof":
		return "Invalid value"
	default:
		return fmt.Sprintf("Failed validation on '%s'", fe.Tag())
	}
}

func NewBadRequestError(msg string) error {
	return &ServiceError{
		Code:    http.StatusBadRequest,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewUnauthorizedError(msg string) error {
	return &ServiceError{
		Code:    http.StatusUnauthorized,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewNotFoundError(msg string) error {
	return &ServiceError{
		Code:    http.StatusNotFound,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewDuplicateError(msg string) error {
	return &ServiceError{
		Code:    http.StatusConflict,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func NewUnprocessableEntityError(msg string) error {
	return &ServiceError{
		Code:    http.StatusUnprocessableEntity,
		Message: msg,
		Err:     errors.New(msg),
	}
}


func NewInternalServerError(msg string) error {
	return &ServiceError{
		Code:    http.StatusInternalServerError,
		Message: msg,
		Err:     errors.New(msg),
	}
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// Check if there are any errors
		if len(c.Errors) > 0 {
			for _, e := range c.Errors {
				// Check if it's a ServiceError
				if serviceErr, ok := e.Err.(*ServiceError); ok {
					if serviceErr.Fields != nil {
						c.JSON(serviceErr.Code, gin.H{
							"error":  serviceErr.Message,
							"fields": serviceErr.Fields,
						})
						return
					}
					c.JSON(serviceErr.Code, gin.H{
						"error": serviceErr.Message,
					})
					return
				}
			}

			// Fallback to generic internal server error
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Internal server error",
			})
		}
	}
}