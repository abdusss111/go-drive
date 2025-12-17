package errors

import (
	"time"

	"github.com/gin-gonic/gin"
)

type ErrorResponse struct {
	Error     string    `json:"error"`
	Code      int       `json:"code"`
	Timestamp time.Time `json:"timestamp"`
}

// HandleError — вспомогательная функция для отправки ошибки
func HandleError(c *gin.Context, err error, statusCode int) {
	c.JSON(statusCode, ErrorResponse{
		Error:     err.Error(),
		Code:      statusCode,
		Timestamp: time.Now(),
	})
}

// HandleErrorWithMessage — отправить ошибку с кастомным сообщением
func HandleErrorWithMessage(c *gin.Context, message string, statusCode int) {
	c.JSON(statusCode, ErrorResponse{
		Error:     message,
		Code:      statusCode,
		Timestamp: time.Now(),
	})
}
