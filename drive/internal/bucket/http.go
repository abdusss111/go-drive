package bucket

import (
	"net/http"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/errors"
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
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	var req createBucketRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		errors.HandleError(c, err, http.StatusBadRequest)
		return
	}

	bucket, err := h.service.CreateBucket(c.Request.Context(), userID, req.Name, req.Description)
	if err != nil {
		switch err {
		case ErrBucketNameExists:
			errors.HandleErrorWithMessage(c, "bucket name already exists", http.StatusConflict)
		default:
			errors.HandleErrorWithMessage(c, "failed to create bucket", http.StatusInternalServerError)
		}
		return
	}

	c.JSON(http.StatusCreated, bucket)
}

func (h *httpHandler) listBuckets(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	buckets, err := h.service.ListBuckets(c.Request.Context(), userID)
	if err != nil {
		errors.HandleErrorWithMessage(c, "failed to list buckets", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"buckets": buckets})
}

func (h *httpHandler) getBucket(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid bucket id", http.StatusBadRequest)
		return
	}

	bucket, err := h.service.GetBucket(c.Request.Context(), userID, bucketID)
	if err != nil {
		if err == ErrBucketNotFound {
			errors.HandleErrorWithMessage(c, "bucket not found", http.StatusNotFound)
			return
		}
		errors.HandleErrorWithMessage(c, "failed to fetch bucket", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, bucket)
}

func (h *httpHandler) deleteBucket(c *gin.Context) {
	userID, _, ok := auth.RequireUser(c)
	if !ok {
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid bucket id", http.StatusBadRequest)
		return
	}

	if err := h.service.DeleteBucket(c.Request.Context(), userID, bucketID); err != nil {
		switch err {
		case ErrBucketNotFound:
			errors.HandleErrorWithMessage(c, "bucket not found", http.StatusNotFound)
		default:
			errors.HandleErrorWithMessage(c, "failed to delete bucket", http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
