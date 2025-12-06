package presigned

import (
	"context"
	"fmt"
	"net/url"
	"time"

	"github.com/google/uuid"
)

var ErrInvalidMethod = fmt.Errorf("invalid method: must be GET or PUT")

type MinioClient interface {
	PresignedGetObject(ctx context.Context, bucket, object string, expiry time.Duration, params map[string]string) (*url.URL, error)
	PresignedPutObject(ctx context.Context, bucket, object string, expiry time.Duration) (*url.URL, error)
}

type Service struct {
	client MinioClient
	ttl    time.Duration
	repo   *Repository
}

func NewService(client MinioClient, ttl time.Duration, repo *Repository) *Service {
	return &Service{
		client: client,
		ttl:    ttl,
		repo:   repo,
	}
}

func (s *Service) GenerateURL(
	ctx context.Context,
	bucketName string,
	objectName string,
	method string,
	userID uuid.UUID,
	bucketID uuid.UUID,
	fileID uuid.UUID,
) (string, Record, AuditRecord, error) {

	if method != "GET" && method != "PUT" {
		return "", Record{}, AuditRecord{}, ErrInvalidMethod
	}

	expiry := s.ttl
	var urlStr string

	switch method {
	case "GET":
		u, err := s.client.PresignedGetObject(ctx, bucketName, objectName, expiry, nil)
		if err != nil {
			return "", Record{}, AuditRecord{}, err
		}
		urlStr = u.String()

	case "PUT":
		u, err := s.client.PresignedPutObject(ctx, bucketName, objectName, expiry)
		if err != nil {
			return "", Record{}, AuditRecord{}, err
		}
		urlStr = u.String()
	}

	rec := Record{
		ID:       uuid.New(),
		ObjectID: fileID,
		Method:   method,
		Expires:  time.Now().Add(expiry),
	}

	saveErr := s.repo.SaveRecord(ctx, rec)
	if saveErr != nil {
		return "", Record{}, AuditRecord{}, saveErr
	}

	audit := AuditRecord{
		ID:        uuid.New(),
		UserID:    userID,
		BucketID:  bucketID,
		FileID:    fileID,
		Method:    method,
		ExpiresAt: rec.Expires,
		CreatedAt: time.Now(),
	}

	auditErr := s.repo.SaveAudit(ctx, audit)
	if auditErr != nil {
		return "", Record{}, AuditRecord{}, auditErr
	}

	return urlStr, rec, audit, nil
}
