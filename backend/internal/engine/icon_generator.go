package engine

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg" // 支持 JPEG 格式解码
	"image/png"
	"math"
	"os"
	"path/filepath"

	"golang.org/x/image/draw"
)

// pwaIconSet 包含生成的所有 PWA 图标路径
type pwaIconSet struct {
	Icon180     string // 180x180 圆角图标（iOS apple-touch-icon）
	Icon192     string // 192x192 圆角图标
	Icon512     string // 512x512 圆角图标
	Maskable192 string // 192x192 maskable 图标（Android 自适应图标）
	Maskable512 string // 512x512 maskable 图标
}

// generateAllPwaIcons 从头像生成完整的 PWA 图标集
// 包括圆角图标（180/192/512）和 maskable 图标（192/512）
func generateAllPwaIcons(appDir, buildDir string, bgColor color.Color) (*pwaIconSet, error) {
	avatarPath := filepath.Join(appDir, "images", "avatar.png")
	if _, err := os.Stat(avatarPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("头像文件不存在: %s", avatarPath)
	}

	src, err := loadImage(avatarPath)
	if err != nil {
		return nil, fmt.Errorf("读取头像失败: %w", err)
	}

	iconsDir := filepath.Join(buildDir, "images", "icons")
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建图标目录失败: %w", err)
	}

	result := &pwaIconSet{}

	// 生成圆角图标（iOS/通用）
	roundedSizes := []struct {
		size int
		dest *string
	}{
		{180, &result.Icon180},
		{192, &result.Icon192},
		{512, &result.Icon512},
	}

	for _, s := range roundedSizes {
		resized := resizeImage(src, s.size, s.size)
		rounded := applyRoundedCorners(resized, s.size)
		filename := fmt.Sprintf("icon-%dx%d.png", s.size, s.size)
		if err := savePNG(filepath.Join(iconsDir, filename), rounded); err != nil {
			return nil, fmt.Errorf("保存图标 %s 失败: %w", filename, err)
		}
		*s.dest = "/images/icons/" + filename
	}

	// 生成 maskable 图标（Android 自适应图标）
	// 核心内容在中心 80% 安全区域，四边各留 10% padding，纯色背景填充
	maskableSizes := []struct {
		size int
		dest *string
	}{
		{192, &result.Maskable192},
		{512, &result.Maskable512},
	}

	for _, s := range maskableSizes {
		maskable := generateMaskableIcon(src, s.size, bgColor)
		filename := fmt.Sprintf("icon-maskable-%dx%d.png", s.size, s.size)
		if err := savePNG(filepath.Join(iconsDir, filename), maskable); err != nil {
			return nil, fmt.Errorf("保存 maskable 图标 %s 失败: %w", filename, err)
		}
		*s.dest = "/images/icons/" + filename
	}

	return result, nil
}

// generateMaskableIcon 生成 Android maskable 图标
// 将头像缩放到安全区域（中心 80%），四周用背景色填充
func generateMaskableIcon(src image.Image, size int, bgColor color.Color) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))

	// 用背景色填充整个画布
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dst.Set(x, y, bgColor)
		}
	}

	// 头像放在中心 80% 安全区域
	padding := int(float64(size) * 0.1) // 四边各 10%
	safeSize := size - 2*padding

	// 将头像缩放到安全区域大小
	resized := resizeImage(src, safeSize, safeSize)

	// 绘制到中心
	for y := 0; y < safeSize; y++ {
		for x := 0; x < safeSize; x++ {
			dst.Set(x+padding, y+padding, resized.At(x, y))
		}
	}

	return dst
}

// generateAllPwaIconsFromSource 从指定图片文件名生成完整 PWA 图标集
func generateAllPwaIconsFromSource(appDir, buildDir, sourceFileName string, bgColor color.Color) (*pwaIconSet, error) {
	sourcePath := filepath.Join(appDir, "images", sourceFileName)
	if _, err := os.Stat(sourcePath); os.IsNotExist(err) {
		return nil, fmt.Errorf("图标文件不存在: %s", sourcePath)
	}

	src, err := loadImage(sourcePath)
	if err != nil {
		return nil, fmt.Errorf("读取图标失败: %w", err)
	}

	iconsDir := filepath.Join(buildDir, "images", "icons")
	if err := os.MkdirAll(iconsDir, 0755); err != nil {
		return nil, fmt.Errorf("创建图标目录失败: %w", err)
	}

	result := &pwaIconSet{}

	// 生成圆角图标
	roundedSizes := []struct {
		size int
		dest *string
	}{
		{180, &result.Icon180},
		{192, &result.Icon192},
		{512, &result.Icon512},
	}

	for _, s := range roundedSizes {
		resized := resizeImage(src, s.size, s.size)
		rounded := applyRoundedCorners(resized, s.size)
		filename := fmt.Sprintf("icon-%dx%d.png", s.size, s.size)
		if err := savePNG(filepath.Join(iconsDir, filename), rounded); err != nil {
			return nil, fmt.Errorf("保存图标 %s 失败: %w", filename, err)
		}
		*s.dest = "/images/icons/" + filename
	}

	// 生成 maskable 图标
	maskableSizes := []struct {
		size int
		dest *string
	}{
		{192, &result.Maskable192},
		{512, &result.Maskable512},
	}

	for _, s := range maskableSizes {
		maskable := generateMaskableIcon(src, s.size, bgColor)
		filename := fmt.Sprintf("icon-maskable-%dx%d.png", s.size, s.size)
		if err := savePNG(filepath.Join(iconsDir, filename), maskable); err != nil {
			return nil, fmt.Errorf("保存 maskable 图标 %s 失败: %w", filename, err)
		}
		*s.dest = "/images/icons/" + filename
	}

	return result, nil
}

// generatePwaIcons 兼容旧调用：生成 192/512 圆角图标
func generatePwaIcons(appDir, buildDir string) (string, string, error) {
	icons, err := generateAllPwaIcons(appDir, buildDir, color.White)
	if err != nil {
		return "", "", err
	}
	return icons.Icon192, icons.Icon512, nil
}

// loadImage 加载图片文件（支持 PNG、JPEG）
func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		return nil, err
	}
	return img, nil
}

// resizeImage 使用 CatmullRom 算法缩放图片到指定尺寸
func resizeImage(src image.Image, width, height int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, width, height))
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)
	return dst
}

// applyRoundedCorners 为正方形图片应用 iOS/Android 风格的连续圆角（squircle）
// 圆角半径约为边长的 22.37%，与 iOS App Icon 规范一致
func applyRoundedCorners(src *image.NRGBA, size int) *image.NRGBA {
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	radius := float64(size) * 0.2237

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			if isInsideRoundedRect(float64(x), float64(y), float64(size), float64(size), radius) {
				dst.Set(x, y, src.At(x, y))
			} else {
				dst.Set(x, y, color.Transparent)
			}
		}
	}

	return dst
}

// isInsideRoundedRect 判断点 (px, py) 是否在圆角矩形内
func isInsideRoundedRect(px, py, w, h, r float64) bool {
	cx := px + 0.5
	cy := py + 0.5

	if cx < r && cy < r {
		return isInsideCorner(cx, cy, r, r, r)
	}
	if cx > w-r && cy < r {
		return isInsideCorner(cx, cy, w-r, r, r)
	}
	if cx < r && cy > h-r {
		return isInsideCorner(cx, cy, r, h-r, r)
	}
	if cx > w-r && cy > h-r {
		return isInsideCorner(cx, cy, w-r, h-r, r)
	}

	return true
}

// isInsideCorner 使用超椭圆（squircle）公式判断点是否在圆角内
func isInsideCorner(px, py, cx, cy, r float64) bool {
	dx := math.Abs(px-cx) / r
	dy := math.Abs(py-cy) / r
	return math.Pow(dx, 5)+math.Pow(dy, 5) <= 1.0
}

// savePNG 保存图片为 PNG 文件
func savePNG(path string, img image.Image) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	return png.Encode(f, img)
}

// parseHexColor 解析 hex 颜色字符串为 color.Color
func parseHexColor(hex string) color.Color {
	if len(hex) == 0 {
		return color.White
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	if len(hex) == 3 {
		hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
	}
	if len(hex) != 6 {
		return color.White
	}

	r := hexByte(hex[0])<<4 | hexByte(hex[1])
	g := hexByte(hex[2])<<4 | hexByte(hex[3])
	b := hexByte(hex[4])<<4 | hexByte(hex[5])

	return color.NRGBA{R: r, G: g, B: b, A: 255}
}

func hexByte(b byte) byte {
	switch {
	case b >= '0' && b <= '9':
		return b - '0'
	case b >= 'a' && b <= 'f':
		return b - 'a' + 10
	case b >= 'A' && b <= 'F':
		return b - 'A' + 10
	default:
		return 0
	}
}
