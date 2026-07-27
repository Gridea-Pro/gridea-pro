package facade

import (
	"context"
	"gridea-pro/backend/internal/service"
	"sync/atomic"
)

type CdnUploadFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.CdnUploadService]
}

func NewCdnUploadFacade(s *service.CdnUploadService) *CdnUploadFacade {
	f := &CdnUploadFacade{}
	f.internal.Store(s)
	return f
}

func (f *CdnUploadFacade) svc() *service.CdnUploadService { return f.internal.Load() }

// TestCdnUpload 测试上传，返回 CDN 访问 URL
func (f *CdnUploadFacade) TestCdnUpload() (string, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().TestUpload(ctx)
}
