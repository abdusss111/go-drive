package presigned

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	bucketModel "github.com/abduss/godrive/internal/bucket"
	fileModel "github.com/abduss/godrive/internal/file"
)

type BucketRepository interface {
	GetBucketByID(ctx context.Context, bucketID uuid.UUID) (bucketModel.Bucket, error)
}

type FileRepository interface {
	GetFileByID(ctx context.Context, id uuid.UUID) (fileModel.Metadata, error)
}

type Handler struct {
	presignedService *Service
	bucketRepo       BucketRepository
	fileRepo         FileRepository
}

func NewHandler(ps *Service, bucketRepo BucketRepository, fileRepo FileRepository) *Handler {
	return &Handler{
		presignedService: ps,
		bucketRepo:       bucketRepo,
		fileRepo:         fileRepo,
	}
}

func (h *Handler) RegisterRoutes(rg *gin.RouterGroup) {
	rg.POST("/buckets/:bucketID/files/:fileID/presigned-url", h.GeneratePresignedURL)
}

func (h *Handler) GeneratePresignedURL(c *gin.Context) {
	bucketID := c.Param("bucketID")
	fileID := c.Param("fileID")
	userID := c.GetString("userID")

	method := c.DefaultQuery("method", "GET")
	ttlParam := c.Query("ttl")

	var ttl time.Duration
	var err error

	if ttlParam != "" {
		ttl, err = time.ParseDuration(ttlParam)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid ttl"})
			return
		}
	} else {
		ttl = h.presignedService.ttl
	}

	bucketUUID, err := uuid.Parse(bucketID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid bucketID"})
		return
	}

	fileUUID, err := uuid.Parse(fileID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid fileID"})
		return
	}

	bucket, err := h.bucketRepo.GetBucketByID(c.Request.Context(), bucketUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		return
	}

	file, err := h.fileRepo.GetFileByID(c.Request.Context(), fileUUID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		return
	}

	if bucket.OwnerID.String() != userID {
		c.JSON(http.StatusForbidden, gin.H{"error": "no access to bucket"})
		return
	}

	if file.BucketID != bucket.ID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file does not belong to this bucket"})
		return
	}

	isPut := method == "PUT"

	url, err := h.presignedService.GeneratePresignedWithAccessCheck(
		c.Request.Context(),
		bucket.Name,
		file.ObjectName,
		userID,
		bucket.OwnerID.String(),
		nil,
		isPut,
	)
	if err != nil {
		c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"url":     url,
		"method":  method,
		"expires": time.Now().Add(ttl),
	})
}
