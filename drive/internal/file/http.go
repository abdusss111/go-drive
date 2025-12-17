package file

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/abduss/godrive/internal/auth"
	"github.com/abduss/godrive/internal/errors"
	"github.com/abduss/godrive/internal/usage"
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
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid bucket id", http.StatusBadRequest)
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		errors.HandleErrorWithMessage(c, "file field is required", http.StatusBadRequest)
		return
	}

	meta, err := h.service.Upload(c.Request.Context(), userID, bucketID, fileHeader)
	if err != nil {
		switch err {
		case ErrBucketMismatch:
			errors.HandleErrorWithMessage(c, "bucket not found", http.StatusNotFound)
		case ErrFileTooLarge:
			errors.HandleErrorWithMessage(c, "file too large", http.StatusBadRequest)
		case usage.ErrQuotaBytesExceeded:
			errors.HandleErrorWithMessage(c, "storage quota exceeded", http.StatusRequestEntityTooLarge)
		case usage.ErrQuotaFilesExceeded:
			errors.HandleErrorWithMessage(c, "file count quota exceeded", http.StatusTooManyRequests)
		case usage.ErrQuotaExceeded:
			errors.HandleErrorWithMessage(c, "quota exceeded", http.StatusTooManyRequests)
		default:
			// Check if it's a quota error by error message
			if strings.Contains(err.Error(), "quota") {
				if strings.Contains(err.Error(), "bytes") {
					errors.HandleErrorWithMessage(c, "storage quota exceeded", http.StatusRequestEntityTooLarge)
				} else {
					errors.HandleErrorWithMessage(c, "quota exceeded", http.StatusTooManyRequests)
				}
			} else {
				errors.HandleErrorWithMessage(c, "failed to upload file", http.StatusInternalServerError)
			}
		}
		return
	}

	c.JSON(http.StatusCreated, meta)
}

func (h *httpHandler) listFiles(c *gin.Context) {
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

	list, err := h.service.List(c.Request.Context(), userID, bucketID)
	if err != nil {
		if err == ErrBucketMismatch {
			errors.HandleErrorWithMessage(c, "bucket not found", http.StatusNotFound)
			return
		}
		errors.HandleErrorWithMessage(c, "failed to list files", http.StatusInternalServerError)
		return
	}

	c.JSON(http.StatusOK, gin.H{"files": list})
}

func (h *httpHandler) downloadFile(c *gin.Context) {
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
	fileID, err := uuid.Parse(c.Param("fileID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid file id", http.StatusBadRequest)
		return
	}

	meta, reader, err := h.service.Download(c.Request.Context(), userID, bucketID, fileID)
	if err != nil {
		switch err {
		case ErrFileNotFound:
			errors.HandleErrorWithMessage(c, "file not found", http.StatusNotFound)
		default:
			errors.HandleErrorWithMessage(c, "failed to download file", http.StatusInternalServerError)
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
		errors.HandleErrorWithMessage(c, "unauthorized", http.StatusUnauthorized)
		return
	}

	bucketID, err := uuid.Parse(c.Param("bucketID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid bucket id", http.StatusBadRequest)
		return
	}
	fileID, err := uuid.Parse(c.Param("fileID"))
	if err != nil {
		errors.HandleErrorWithMessage(c, "invalid file id", http.StatusBadRequest)
		return
	}

	if err := h.service.Delete(c.Request.Context(), userID, bucketID, fileID); err != nil {
		switch err {
		case ErrFileNotFound:
			errors.HandleErrorWithMessage(c, "file not found", http.StatusNotFound)
		case ErrBucketMismatch:
			errors.HandleErrorWithMessage(c, "bucket not found", http.StatusNotFound)
		default:
			errors.HandleErrorWithMessage(c, "failed to delete file", http.StatusInternalServerError)
		}
		return
	}

	c.Status(http.StatusNoContent)
}
