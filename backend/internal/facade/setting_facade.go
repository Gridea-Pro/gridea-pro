package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// SettingFacade wraps SettingService
type SettingFacade struct {
	internal *service.SettingService
}

func NewSettingFacade(s *service.SettingService) *SettingFacade {
	return &SettingFacade{internal: s}
}

func (f *SettingFacade) GetSetting() (domain.Setting, error) {
	return f.internal.GetSetting(context.TODO())
}

func (f *SettingFacade) SaveAvatar(sourcePath string) error {
	return f.internal.SaveAvatar(sourcePath)
}

func (f *SettingFacade) SaveFavicon(sourcePath string) error {
	return f.internal.SaveFavicon(sourcePath)
}
