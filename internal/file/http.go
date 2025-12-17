package file

import (
	"fmt"
	"io"
	"net/http"

	"github.com/abduss/godrive/internal/auth"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// RegisterRoutes mounts file operations under the provided router group.
func RegisterRoutes(group *gin.RouterGroup, service *Service) {
	handler := &httpHandler{service: service}
	group.POST("/buckets/:bucketID/files", handler.uploadFile)
	group.GET("/buckets/:bucketID/files", handler.listFiles)
	group.GET("/buckets/:bucketID/files/:fileID/download", handler.downloadFile)
	group.DELETE("/buckets/:bucketID/files/:fileID", handler.deleteFile)
}

type httpHandler struct {
	service *Service
}

func (h *httpHandler) uploadFile(c *gin.Context) {
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

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "file field is required"})
		return
	}

	meta, err := h.service.Upload(c.Request.Context(), userID, bucketID, fileHeader)
	if err != nil {
		switch err {
		case ErrBucketMismatch:
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		case ErrFileTooLarge:
			c.JSON(http.StatusBadRequest, gin.H{"error": "file too large"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to upload file"})
		}
		return
	}

	c.JSON(http.StatusCreated, meta)
}

func (h *httpHandler) listFiles(c *gin.Context) {
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

	list, err := h.service.List(c.Request.Context(), userID, bucketID)
	if err != nil {
		if err == ErrBucketMismatch {
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list files"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": list})
}

func (h *httpHandler) downloadFile(c *gin.Context) {
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

	meta, reader, err := h.service.Download(c.Request.Context(), userID, bucketID, fileID)
	if err != nil {
		switch err {
		case ErrFileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to download file"})
		}
		return
	}
	defer reader.Close()

	c.Header("Content-Type", meta.ContentType)
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%q", meta.OriginalFilename))
	c.Header("Content-Length", fmt.Sprintf("%d", meta.SizeBytes))

	if _, err := io.Copy(c.Writer, reader); err != nil {
		c.Status(http.StatusInternalServerError)
		return
	}
}

func (h *httpHandler) deleteFile(c *gin.Context) {
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

	if err := h.service.Delete(c.Request.Context(), userID, bucketID, fileID); err != nil {
		switch err {
		case ErrFileNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
		case ErrBucketMismatch:
			c.JSON(http.StatusNotFound, gin.H{"error": "bucket not found"})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete file"})
		}
		return
	}

	c.Status(http.StatusNoContent)
}
