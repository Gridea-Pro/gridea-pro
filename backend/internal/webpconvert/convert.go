package webpconvert

import (
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"os"
	"path/filepath"
	"strings"

	"github.com/chai2010/webp"
)

// Package webpconvert 提供图片转 WebP 的独立转换能力。
// 转换结果写入临时文件，调用方负责复制到最终目的地并清理临时文件。

// convertibleExts 是可以转换为 WebP 的图片格式扩展名。
var convertibleExts = map[string]bool{
	".jpg":  true,
	".jpeg": true,
	".png":  true,
	".bmp":  true,
}

// skipExts 是不需要转换的格式（已是 WebP 或不支持转换）。
var skipExts = map[string]bool{
	".webp": true,
	".svg":  true,
	".ico":  true,
	".gif":  true, // 动画 GIF 只能读取第一帧，转 WebP 会静默丢失动画
}

// ConvertToWebP 将图片文件转换为 WebP 格式。
// 返回临时文件路径（调用方负责复制到最终目的地并清理）。
// 若图片已是 WebP 或格式不支持转换，返回原路径且 tmpPath 为空。
func ConvertToWebP(srcPath string, quality int) (tmpPath string, err error) {
	ext := strings.ToLower(filepath.Ext(srcPath))

	if skipExts[ext] {
		// 已是 WebP 或不支持的格式，跳过
		return "", nil
	}

	if !convertibleExts[ext] {
		// 未知格式，跳过
		return "", nil
	}

	if quality <= 0 || quality > 100 {
		quality = 80
	}

	// 打开并解码源图片
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	img, _, err := image.Decode(srcFile)
	if err != nil {
		return "", fmt.Errorf("解码图片失败: %w", err)
	}

	// 编码为 WebP（优先 RGBA，失败时降级 RGB）
	data, err := webp.EncodeRGBA(img, float32(quality))
	if err != nil {
		data, err = webp.EncodeRGB(img, float32(quality))
		if err != nil {
			return "", fmt.Errorf("WebP 编码失败: %w", err)
		}
	}

	// 写入临时文件
	tmpFile, err := os.CreateTemp("", "gridea-webp-*.webp")
	if err != nil {
		return "", fmt.Errorf("创建临时文件失败: %w", err)
	}
	defer tmpFile.Close()

	if _, err := tmpFile.Write(data); err != nil {
		os.Remove(tmpFile.Name())
		return "", fmt.Errorf("写入临时文件失败: %w", err)
	}

	return tmpFile.Name(), nil
}

// NeedsConversion 判断文件是否需要 WebP 转换。
func NeedsConversion(filePath string) bool {
	ext := strings.ToLower(filepath.Ext(filePath))
	return convertibleExts[ext]
}
