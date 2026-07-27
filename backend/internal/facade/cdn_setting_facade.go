package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
)

type CdnSettingFacade struct {
	// repo 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免 data race。
	repo atomic.Value // 存 domain.CdnSettingRepository
}

func NewCdnSettingFacade(repo domain.CdnSettingRepository) *CdnSettingFacade {
	f := &CdnSettingFacade{}
	f.repo.Store(repo)
	return f
}

func (f *CdnSettingFacade) repository() domain.CdnSettingRepository {
	return f.repo.Load().(domain.CdnSettingRepository)
}

func (f *CdnSettingFacade) GetCdnSetting() (domain.CdnSetting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().GetCdnSetting(ctx)
}

func (f *CdnSettingFacade) SaveCdnSettingFromFrontend(setting domain.CdnSetting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repository().SaveCdnSetting(ctx, setting)
}
