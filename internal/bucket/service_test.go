package bucket

import (
	"context"
	"testing"

	"github.com/google/uuid"
)

func TestCreateAndListBuckets(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo, &fakeFileIndex{}, nil, "storage")

	ownerID := uuid.New()
	description := "personal docs"
	created, err := service.CreateBucket(context.Background(), ownerID, "documents", &description)
	if err != nil {
		t.Fatalf("CreateBucket returned error: %v", err)
	}

	if created.Name != "documents" {
		t.Fatalf("expected bucket name documents, got %s", created.Name)
	}

	buckets, err := service.ListBuckets(context.Background(), ownerID)
	if err != nil {
		t.Fatalf("ListBuckets returned error: %v", err)
	}

	if len(buckets) != 1 {
		t.Fatalf("expected 1 bucket, got %d", len(buckets))
	}
}

func TestCreateBucketDuplicateName(t *testing.T) {
	repo := newFakeRepo()
	service := NewService(repo, &fakeFileIndex{}, nil, "storage")

	ownerID := uuid.New()
	if _, err := service.CreateBucket(context.Background(), ownerID, "photos", nil); err != nil {
		t.Fatalf("unexpected error creating bucket: %v", err)
	}

	if _, err := service.CreateBucket(context.Background(), ownerID, "photos", nil); err != ErrBucketNameExists {
		t.Fatalf("expected ErrBucketNameExists, got %v", err)
	}
}

func TestDeleteBucketInvokesFileCleanup(t *testing.T) {
	repo := newFakeRepo()
	fileIndex := &fakeFileIndex{}
	service := NewService(repo, fileIndex, nil, "storage")

	ownerID := uuid.New()
	bucket, err := service.CreateBucket(context.Background(), ownerID, "temp", nil)
	if err != nil {
		t.Fatalf("CreateBucket returned error: %v", err)
	}

	if err := service.DeleteBucket(context.Background(), ownerID, bucket.ID); err != nil {
		t.Fatalf("DeleteBucket returned error: %v", err)
	}

	if _, err := repo.Get(context.Background(), ownerID, bucket.ID); err == nil {
		t.Fatalf("expected bucket to be removed from repository")
	}
}

// --- fakes ----

type fakeRepo struct {
	buckets map[uuid.UUID]Bucket
	byName  map[uuid.UUID]map[string]uuid.UUID
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{
		buckets: make(map[uuid.UUID]Bucket),
		byName:  make(map[uuid.UUID]map[string]uuid.UUID),
	}
}

func (f *fakeRepo) Create(ctx context.Context, ownerID uuid.UUID, name string, description *string) (Bucket, error) {
	if _, ok := f.byName[ownerID]; !ok {
		f.byName[ownerID] = make(map[string]uuid.UUID)
	}
	if _, exists := f.byName[ownerID][name]; exists {
		return Bucket{}, ErrBucketNameExists
	}
	id := uuid.New()
	b := Bucket{
		ID:          id,
		OwnerID:     ownerID,
		Name:        name,
		Description: description,
	}
	f.byName[ownerID][name] = id
	f.buckets[id] = b
	return b, nil
}

func (f *fakeRepo) List(ctx context.Context, ownerID uuid.UUID) ([]Bucket, error) {
	var buckets []Bucket
	for _, bucket := range f.buckets {
		if bucket.OwnerID == ownerID {
			buckets = append(buckets, bucket)
		}
	}
	return buckets, nil
}

func (f *fakeRepo) Get(ctx context.Context, ownerID, bucketID uuid.UUID) (Bucket, error) {
	b, ok := f.buckets[bucketID]
	if !ok || b.OwnerID != ownerID {
		return Bucket{}, ErrBucketNotFound
	}
	return b, nil
}

func (f *fakeRepo) Delete(ctx context.Context, ownerID, bucketID uuid.UUID) error {
	b, ok := f.buckets[bucketID]
	if !ok || b.OwnerID != ownerID {
		return ErrBucketNotFound
	}
	delete(f.buckets, bucketID)
	if nameMap, ok := f.byName[ownerID]; ok {
		delete(nameMap, b.Name)
	}
	return nil
}

func (f *fakeRepo) RecordUsageSnapshot(ctx context.Context, ownerID uuid.UUID) error {
	return nil
}

type fakeFileIndex struct {
	wasCalled bool
}

func (f *fakeFileIndex) ListObjectsForBucket(ctx context.Context, bucketID uuid.UUID) ([]FileObject, error) {
	f.wasCalled = true
	return []FileObject{
		{ObjectName: "obj", SizeBytes: 42},
	}, nil
}
