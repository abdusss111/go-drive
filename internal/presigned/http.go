package presigned

import (
	"net/http"
	"time"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/file"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes mounts presigned URL endpoints.
func RegisterRoutes(group *gin.RouterGroup, service *Service) {
	handler := &httpHandler{service: service}
	group.POST("/buckets/:bucketID/files/:fileID/presigned-url", handler.createPresignedURL)
}

type httpHandler struct {
	service *Service
}

func (h *httpHandler) createPresignedURL(c *gin.Context) {
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

	method := c.DefaultQuery("method", http.MethodGet)

	ttl := h.parseTTL(c.Query("ttl"))

	resp, err := h.service.Generate(c.Request.Context(), GenerateRequest{
		UserID:   userID,
		BucketID: bucketID,
		FileID:   fileID,
		Method:   method,
		TTL:      ttl,
	})
	if err != nil {
		switch err {
		case bucket.ErrBucketNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		case file.ErrFileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		}
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *httpHandler) parseTTL(raw string) time.Duration {
	if raw == "" {
		return 0
	}
	dur, err := time.ParseDuration(raw)
	if err != nil {
		return 0
	}
	return dur
}
