package presigned

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/abduss/godrive/internal/bucket"
	"github.com/abduss/godrive/internal/file"
	"github.com/google/uuid"
)

// Repository persists presigned URL audit entries.
type Repository interface {
	Log(ctx context.Context, entry AuditEntry) error
}

// AuditEntry represents a single presigned URL creation event.
type AuditEntry struct {
	UserID    uuid.UUID
	BucketID  uuid.UUID
	FileID    uuid.UUID
	Method    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

// Service generates presigned URLs with ownership validation and audit logging.
type Service struct {
	repo             Repository
	files            fileMetadataProvider
	buckets          bucketProvider
	objectStore      presigner
	objectBucket     string
	defaultTTL       time.Duration
	maxTTL           time.Duration
	scopeTokenSecret string
}

// fileMetadataProvider defines the subset of file.Repository used by the service.
type fileMetadataProvider interface {
	Get(ctx context.Context, ownerID, bucketID, fileID uuid.UUID) (file.Metadata, error)
}

// bucketProvider defines the subset of bucket.Repository used by the service.
type bucketProvider interface {
	Get(ctx context.Context, ownerID, bucketID uuid.UUID) (bucket.Bucket, error)
}

type presigner interface {
	PresignedGetObject(ctx context.Context, bucketName, objectName string, expiry time.Duration, reqParams url.Values) (*url.URL, error)
	PresignedPutObject(ctx context.Context, bucketName, objectName string, expiry time.Duration) (*url.URL, error)
}

// NewService constructs a presigned URL service.
func NewService(repo Repository, files fileMetadataProvider, buckets bucketProvider, objectStore presigner, objectBucket string, defaultTTL, maxTTL time.Duration, scopeSecret string) *Service {
	return &Service{
		repo:             repo,
		files:            files,
		buckets:          buckets,
		objectStore:      objectStore,
		objectBucket:     objectBucket,
		defaultTTL:       defaultTTL,
		maxTTL:           maxTTL,
		scopeTokenSecret: scopeSecret,
	}
}

// Generate builds a presigned URL with access checks and audit logging.
func (s *Service) Generate(ctx context.Context, req GenerateRequest) (GenerateResponse, error) {
	method := strings.ToUpper(strings.TrimSpace(req.Method))
	if method == "" {
		method = httpMethodGet
	}
	if method != httpMethodGet && method != httpMethodPut {
		return GenerateResponse{}, fmt.Errorf("unsupported method %s", method)
	}

	ttl := req.TTL
	if ttl <= 0 {
		ttl = s.defaultTTL
	}
	if s.maxTTL > 0 && ttl > s.maxTTL {
		ttl = s.maxTTL
	}
	if ttl <= 0 {
		return GenerateResponse{}, fmt.Errorf("invalid ttl")
	}

	if _, err := s.buckets.Get(ctx, req.UserID, req.BucketID); err != nil {
		return GenerateResponse{}, translateBucketError(err)
	}

	meta, err := s.files.Get(ctx, req.UserID, req.BucketID, req.FileID)
	if err != nil {
		return GenerateResponse{}, translateFileError(err)
	}

	var (
		url    string
		expiry = ttl
	)

	switch method {
	case httpMethodGet:
		url, err = s.generateGetURL(ctx, meta.ObjectName, expiry)
	case httpMethodPut:
		url, err = s.generatePutURL(ctx, meta.ObjectName, expiry)
	}
	if err != nil {
		return GenerateResponse{}, err
	}

	expiresAt := time.Now().Add(expiry)
	scopeToken := s.makeScopeToken(req.BucketID, req.FileID, method, expiresAt)

	if s.repo != nil {
		_ = s.repo.Log(ctx, AuditEntry{
			UserID:    req.UserID,
			BucketID:  req.BucketID,
			FileID:    req.FileID,
			Method:    method,
			ExpiresAt: expiresAt,
			CreatedAt: time.Now(),
		})
	}

	return GenerateResponse{
		URL:        url,
		Method:     method,
		ExpiresAt:  expiresAt,
		ScopeToken: scopeToken,
	}, nil
}

func (s *Service) generateGetURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presigned, err := s.objectStore.PresignedGetObject(ctx, s.objectBucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign GET: %w", err)
	}
	return presigned.String(), nil
}

func (s *Service) generatePutURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	presigned, err := s.objectStore.PresignedPutObject(ctx, s.objectBucket, objectName, expiry)
	if err != nil {
		return "", fmt.Errorf("presign PUT: %w", err)
	}
	return presigned.String(), nil
}

func (s *Service) makeScopeToken(bucketID, fileID uuid.UUID, method string, expiresAt time.Time) string {
	if s.scopeTokenSecret == "" {
		return ""
	}
	mac := hmac.New(sha256.New, []byte(s.scopeTokenSecret))
	data := fmt.Sprintf("%s|%s|%s|%d", bucketID.String(), fileID.String(), method, expiresAt.Unix())
	_, _ = mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}

const (
	httpMethodGet = "GET"
	httpMethodPut = "PUT"
)

func translateBucketError(err error) error {
	switch err {
	case bucket.ErrBucketNotFound:
		return err
	default:
		return fmt.Errorf("bucket validation failed: %w", err)
	}
}

func translateFileError(err error) error {
	switch err {
	case file.ErrFileNotFound:
		return err
	default:
		return fmt.Errorf("file validation failed: %w", err)
	}
}
