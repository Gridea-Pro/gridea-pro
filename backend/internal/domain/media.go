package domain

import "context"

type MediaRepository interface {
	SaveImages(ctx context.Context, files []UploadedFile) ([]string, error)
	// SaveImageBytes 保存内存中的图片字节（粘贴 / 拖拽场景），返回相对路径 /post-images/xxx
	SaveImageBytes(ctx context.Context, name string, data []byte) (string, error)
}
