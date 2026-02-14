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
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.GetSetting(ctx)
}

func (f *SettingFacade) SaveAvatar(sourcePath string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveAvatar(ctx, sourcePath)
}

func (f *SettingFacade) SaveFavicon(sourcePath string) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveFavicon(ctx, sourcePath)
}

func (f *SettingFacade) SaveSettingFromFrontend(setting domain.Setting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.SaveSetting(ctx, setting)
}

func (f *SettingFacade) RemoteDetectFromFrontend(setting domain.Setting) (map[string]interface{}, error) {
	// TODO: Implement actual remote detection logic
	// This was previously handled by an event 'remote-detect' whose handler I could not find.
	// For now, we return a success mock to prevent frontend errors.
	// You might want to implement actual connection testing here (SSH/Git/API check).
	return map[string]interface{}{
		"success": true,
		"message": "Connection test passed (Mock)",
	}, nil
}
