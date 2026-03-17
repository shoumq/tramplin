package storage

import (
	"context"
	"fmt"
	"io"
)

type AvatarUploadResult struct {
	ObjectKey string
	URL       string
}

type Storage interface {
	UploadAvatar(ctx context.Context, userID, fileName, contentType string, size int64, body io.Reader) (*AvatarUploadResult, error)
}

type NoopStorage struct{}

func (NoopStorage) UploadAvatar(_ context.Context, _, _, _ string, _ int64, _ io.Reader) (*AvatarUploadResult, error) {
	return nil, fmt.Errorf("storage is not configured")
}
