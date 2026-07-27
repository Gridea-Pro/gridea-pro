package domain

import "context"

// ImageOptimizeSetting 图片优化设置
type ImageOptimizeSetting struct {
	Enabled bool `json:"enabled"`
	Quality int  `json:"quality"` // 保留字段，暂固定为 80，后续版本开放可配置
}

// GetQuality 返回压缩质量，未设置时默认 80。
func (s *ImageOptimizeSetting) GetQuality() int {
	if s.Quality <= 0 || s.Quality > 100 {
		return 80
	}
	return s.Quality
}

// ImageOptimizeSettingRepository 定义图片优化设置存储接口
type ImageOptimizeSettingRepository interface {
	GetImageOptimizeSetting(ctx context.Context) (ImageOptimizeSetting, error)
	SaveImageOptimizeSetting(ctx context.Context, setting ImageOptimizeSetting) error
}
