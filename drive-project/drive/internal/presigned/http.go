package presigned

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/file"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes mounts presigned URL endpoints onto the router.
func RegisterRoutes(group *gin.RouterGroup, service *Service) {
	handler := &httpHandler{service: service}
	group.POST("/buckets/:bucketID/files/:fileID/presigned-url", handler.generatePresignedURL)
}

type httpHandler struct {
	service *Service
}

// generatePresignedURL generates a presigned URL for a file.
func (h *httpHandler) generatePresignedURL(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucket id"})
		return
	}

	fileID, err := uuid.Parse(c.Param("fileID"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid file id"})
		return
	}

	var req PresignedURLRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse TTL if provided
	ttl := time.Hour // default
	if req.TTL != "" {
		parsedTTL, err := time.ParseDuration(req.TTL)
		if err == nil && parsedTTL > 0 {
			ttl = parsedTTL
		}
	}

	response, err := h.service.GeneratePresignedURL(c.Request.Context(), userID, bucketID, fileID, req.Method, ttl)
	if err != nil {
		// Check for specific errors using errors.Is for wrapped errors
		switch {
		case errors.Is(err, bucket.ErrBucketNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		case errors.Is(err, file.ErrFileNotFound):
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		default:
			// Fallback: check error message for wrapped errors
			errStr := err.Error()
			if strings.Contains(errStr, "bucket not found") || strings.Contains(errStr, "verify bucket") {
				c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
			} else if strings.Contains(errStr, "file not found") || strings.Contains(errStr, "verify file") {
				c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			} else {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to generate presigned URL"})
			}
		}
		return
	}

	c.JSON(http.StatusOK, response)
}

