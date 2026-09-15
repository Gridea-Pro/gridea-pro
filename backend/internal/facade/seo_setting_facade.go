package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
)

type SeoSettingFacade struct {
	// repo 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免 data race。
	repo atomic.Value // 存 domain.SeoSettingRepository
}

func NewSeoSettingFacade(repo domain.SeoSettingRepository) *SeoSettingFacade {
	f := &SeoSettingFacade{}
	f.repo.Store(repo)
	return f
}

func (f *SeoSettingFacade) repository() domain.SeoSettingRepository {
	return f.repo.Load().(domain.SeoSettingRepository)
}

func (f *SeoSettingFacade) GetSeoSetting() (domain.SeoSetting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().GetSeoSetting(ctx)
}

func (f *SeoSettingFacade) SaveSeoSettingFromFrontend(setting domain.SeoSetting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().SaveSeoSetting(ctx, setting)
}
