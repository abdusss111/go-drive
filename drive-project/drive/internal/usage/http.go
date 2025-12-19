package usage

import (
	"context"
	"net/http"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/bucket"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes mounts usage endpoints onto the router.
func RegisterRoutes(group *gin.RouterGroup, service *Service, bucketRepo bucketRepository) {
	handler := &httpHandler{
		service:    service,
		bucketRepo: bucketRepo,
	}
	group.GET("/usage", handler.getUserUsage)
	group.GET("/buckets/:bucketID/usage", handler.getBucketUsage)
}

type bucketRepository interface {
	GetBucket(ctx context.Context, ownerID, bucketID uuid.UUID) (bucket.Bucket, error)
}

type httpHandler struct {
	service    *Service
	bucketRepo bucketRepository
}

// getUserUsage returns the current user's total usage across all buckets.
// @Summary Get user usage
// @Description Fetch total storage usage and file count across all buckets.
// @Tags Usage
// @Produce json
// @Security Bearer
// @Success 200 {object} UsageResponse
// @Failure 401 {object} map[string]string "unauthorized"
// @Router /usage [get]
func (h *httpHandler) getUserUsage(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	usage, err := h.service.GetUserUsage(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

// getBucketUsage returns the usage statistics for a specific bucket.
// @Summary Get bucket usage
// @Description Fetch storage usage and file count for a specific bucket.
// @Tags Usage
// @Produce json
// @Security Bearer
// @Param bucketID path string true "Bucket UUID"
// @Success 200 {object} BucketUsageResponse
// @Failure 401 {object} map[string]string "unauthorized"
// @Failure 404 {object} map[string]string "bucket not found"
// @Router /buckets/{bucketID}/usage [get]
func (h *httpHandler) getBucketUsage(c *gin.Context) {
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

	// Verify bucket ownership
	_, err = h.bucketRepo.GetBucket(c.Request.Context(), userID, bucketID)
	if err != nil {
		if err == bucket.ErrBucketNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to verify bucket ownership"})
		return
	}

	usage, err := h.service.GetBucketUsage(c.Request.Context(), bucketID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get bucket usage"})
		return
	}

	c.JSON(http.StatusOK, usage)
}

