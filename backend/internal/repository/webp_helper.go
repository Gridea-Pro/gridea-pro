package repository

import (
	"context"
	"gridea-pro/backend/internal/domain"
)

// GetWebpSetting 获取 WebP 转换设置。repo 为 nil 或读取失败时返回关闭状态。
func GetWebpSetting(repo domain.ImageOptimizeSettingRepository, ctx context.Context) (enabled bool, quality int) {
	if repo == nil {
		return false, 80
	}
	setting, err := repo.GetImageOptimizeSetting(ctx)
	if err != nil {
		return false, 80
	}
	return setting.Enabled, setting.GetQuality()
}
