package presigned

import (
	"context"
	"errors"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
)

type ScopeToken struct {
	UserID    string
	Bucket    string
	Object    string
	CanRead   bool
	CanWrite  bool
	ExpiresAt time.Time
}

type Service struct {
	minioClient *minio.Client
	ttl         time.Duration
}

func NewService(minioClient *minio.Client, ttl time.Duration) *Service {
	return &Service{
		minioClient: minioClient,
		ttl:         ttl,
	}
}

func (s *Service) ValidateOwnership(userID, ownerID string) bool {
	return userID == ownerID
}

func (s *Service) ValidateScope(token ScopeToken, bucket, object string, write bool) bool {
	if time.Now().After(token.ExpiresAt) {
		return false
	}
	if token.Bucket != bucket {
		return false
	}
	if token.Object != object {
		return false
	}
	if write && !token.CanWrite {
		return false
	}
	if !write && !token.CanRead {
		return false
	}
	return true
}

func (s *Service) GenerateGetURL(ctx context.Context, bucket, object string) (string, error) {
	reqParams := make(url.Values)

	u, err := s.minioClient.PresignedGetObject(
		ctx,
		bucket,
		object,
		s.ttl,
		reqParams,
	)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (s *Service) GeneratePutURL(ctx context.Context, bucket, object string) (string, error) {
	u, err := s.minioClient.PresignedPutObject(
		ctx,
		bucket,
		object,
		s.ttl,
	)
	if err != nil {
		return "", err
	}

	return u.String(), nil
}

func (s *Service) GeneratePresignedWithAccessCheck(
	ctx context.Context,
	bucket string,
	object string,
	userID string,
	ownerID string,
	scope *ScopeToken,
	write bool,
) (string, error) {

	if scope != nil {
		if !s.ValidateScope(*scope, bucket, object, write) {
			return "", errors.New("access denied: invalid scope token")
		}
	}

	if !s.ValidateOwnership(userID, ownerID) {
		return "", errors.New("access denied: not the owner")
	}

	if write {
		return s.GeneratePutURL(ctx, bucket, object)
	}

	return s.GenerateGetURL(ctx, bucket, object)
}
