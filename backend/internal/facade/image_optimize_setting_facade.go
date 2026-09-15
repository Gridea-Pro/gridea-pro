package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
)

type ImageOptimizeSettingFacade struct {
	// repo 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免 data race。
	repo atomic.Value // 存 domain.ImageOptimizeSettingRepository
}

func NewImageOptimizeSettingFacade(repo domain.ImageOptimizeSettingRepository) *ImageOptimizeSettingFacade {
	f := &ImageOptimizeSettingFacade{}
	f.repo.Store(repo)
	return f
}

func (f *ImageOptimizeSettingFacade) repository() domain.ImageOptimizeSettingRepository {
	return f.repo.Load().(domain.ImageOptimizeSettingRepository)
}

func (f *ImageOptimizeSettingFacade) GetImageOptimizeSetting() (domain.ImageOptimizeSetting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().GetImageOptimizeSetting(ctx)
}

func (f *ImageOptimizeSettingFacade) SaveImageOptimizeSettingFromFrontend(setting domain.ImageOptimizeSetting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().SaveImageOptimizeSetting(ctx, setting)
}
