package bucket

import (
	"net/http"

	"github.com/abduss/godrive/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes mounts bucket endpoints onto the router.
func RegisterRoutes(group *gin.RouterGroup, service *Service) {
	handler := &httpHandler{service: service}
	group.POST("/buckets", handler.createBucket)
	group.GET("/buckets", handler.listBuckets)
	group.GET("/buckets/:bucketID", handler.getBucket)
	group.DELETE("/buckets/:bucketID", handler.deleteBucket)
}

type httpHandler struct {
	service *Service
}

type createBucketRequest struct {
	Name        string  `json:"name" binding:"required"`
	Description *string `json:"description" binding:"omitempty,max=255"`
}

func (h *httpHandler) createBucket(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var req createBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	bucket, err := h.service.CreateBucket(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		switch err {
		case ErrBucketNameExists:
			c.JSON(http.StatusConflict, gin.H{"error": "bucket name already exists"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create bucket"})
		}
		return
	}

	c.JSON(http.StatusCreated, bucket)
}

func (h *httpHandler) listBuckets(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	buckets, err := h.service.ListBuckets(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list buckets"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

func (h *httpHandler) getBucket(c *gin.Context) {
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

	bucket, err := h.service.GetBucket(c.Request.Context(), userID, bucketID)
	if err != nil {
		if err == ErrBucketNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch bucket"})
		return
	}

	c.JSON(http.StatusOK, bucket)
}

func (h *httpHandler) deleteBucket(c *gin.Context) {
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

	if err := h.service.DeleteBucket(c.Request.Context(), userID, bucketID); err != nil {
		switch err {
		case ErrBucketNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete bucket"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
