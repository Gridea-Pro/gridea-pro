package domain

import "context"

type MediaRepository interface {
	SaveImages(ctx context.Context, files []UploadedFile) ([]string, error)
}
