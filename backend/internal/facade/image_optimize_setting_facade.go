package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

type ImageOptimizeSettingFacade struct {
	repo domain.ImageOptimizeSettingRepository
}

func NewImageOptimizeSettingFacade(repo domain.ImageOptimizeSettingRepository) *ImageOptimizeSettingFacade {
	return &ImageOptimizeSettingFacade{repo: repo}
}

func (f *ImageOptimizeSettingFacade) GetImageOptimizeSetting() (domain.ImageOptimizeSetting, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repo.GetImageOptimizeSetting(ctx)
}

func (f *ImageOptimizeSettingFacade) SaveImageOptimizeSettingFromFrontend(setting domain.ImageOptimizeSetting) error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.repo.SaveImageOptimizeSetting(ctx, setting)
}
