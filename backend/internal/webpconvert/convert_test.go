package webpconvert

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// createTestJPEG 创建一个临时 JPEG 测试图片。
func createTestJPEG(t *testing.T) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-*.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x * 2), G: uint8(y * 2), B: 128, A: 255})
		}
	}

	if err := jpeg.Encode(tmpFile, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

// createTestPNG 创建一个临时 PNG 测试图片。
func createTestPNG(t *testing.T) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-*.png")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	for y := 0; y < 100; y++ {
		for x := 0; x < 100; x++ {
			img.Set(x, y, color.RGBA{R: 200, G: uint8(x * 2), B: uint8(y * 2), A: 255})
		}
	}

	if err := png.Encode(tmpFile, img); err != nil {
		t.Fatal(err)
	}
	return tmpFile.Name()
}

// createTestWebP 创建一个临时 WebP 测试文件（模拟已是 WebP 的文件）。
func createTestWebP(t *testing.T) string {
	t.Helper()
	tmpFile, err := os.CreateTemp("", "test-*.webp")
	if err != nil {
		t.Fatal(err)
	}
	defer tmpFile.Close()
	// 写入一些假数据，只要扩展名是 .webp 即可
	tmpFile.Write([]byte("RIFF....WEBP"))
	return tmpFile.Name()
}

func TestConvertToWebP_JPEG(t *testing.T) {
	srcPath := createTestJPEG(t)
	defer os.Remove(srcPath)

	tmpPath, err := ConvertToWebP(srcPath, 80)
	if err != nil {
		t.Fatalf("ConvertToWebP failed: %v", err)
	}
	if tmpPath == "" {
		t.Fatal("expected non-empty tmpPath for JPEG")
	}
	defer os.Remove(tmpPath)

	if !strings.HasSuffix(tmpPath, ".webp") {
		t.Errorf("expected .webp extension, got %s", filepath.Ext(tmpPath))
	}

	// 验证临时文件存在且非空
	info, err := os.Stat(tmpPath)
	if err != nil {
		t.Fatalf("temp file not found: %v", err)
	}
	if info.Size() == 0 {
		t.Error("temp file is empty")
	}
}

func TestConvertToWebP_PNG(t *testing.T) {
	srcPath := createTestPNG(t)
	defer os.Remove(srcPath)

	tmpPath, err := ConvertToWebP(srcPath, 80)
	if err != nil {
		t.Fatalf("ConvertToWebP failed: %v", err)
	}
	if tmpPath == "" {
		t.Fatal("expected non-empty tmpPath for PNG")
	}
	defer os.Remove(tmpPath)

	if !strings.HasSuffix(tmpPath, ".webp") {
		t.Errorf("expected .webp extension, got %s", filepath.Ext(tmpPath))
	}
}

func TestConvertToWebP_AlreadyWebP(t *testing.T) {
	srcPath := createTestWebP(t)
	defer os.Remove(srcPath)

	tmpPath, err := ConvertToWebP(srcPath, 80)
	if err != nil {
		t.Fatalf("ConvertToWebP should not error for WebP: %v", err)
	}
	if tmpPath != "" {
		t.Errorf("expected empty tmpPath for already-WebP, got %s", tmpPath)
		os.Remove(tmpPath)
	}
}

func TestConvertToWebP_SVG(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "test-*.svg")
	if err != nil {
		t.Fatal(err)
	}
	tmpFile.WriteString("<svg></svg>")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	tmpPath, err := ConvertToWebP(tmpFile.Name(), 80)
	if err != nil {
		t.Fatalf("ConvertToWebP should not error for SVG: %v", err)
	}
	if tmpPath != "" {
		t.Errorf("expected empty tmpPath for SVG, got %s", tmpPath)
		os.Remove(tmpPath)
	}
}

func TestConvertToWebP_InvalidFile(t *testing.T) {
	_, err := ConvertToWebP("/nonexistent/file.jpg", 80)
	if err == nil {
		t.Error("expected error for nonexistent file")
	}
}

func TestConvertToWebP_QualityDefault(t *testing.T) {
	srcPath := createTestJPEG(t)
	defer os.Remove(srcPath)

	// quality 0 should default to 80
	tmpPath, err := ConvertToWebP(srcPath, 0)
	if err != nil {
		t.Fatalf("ConvertToWebP failed with quality 0: %v", err)
	}
	if tmpPath == "" {
		t.Fatal("expected non-empty tmpPath")
	}
	defer os.Remove(tmpPath)
}

func TestNeedsConversion(t *testing.T) {
	tests := []struct {
		path string
		want bool
	}{
		{"image.jpg", true},
		{"image.jpeg", true},
		{"image.png", true},
		{"image.gif", false}, // 动画 GIF 不转换，避免丢失动画
		{"image.bmp", true},
		{"image.webp", false},
		{"image.svg", false},
		{"image.ico", false},
		{"image.JPG", true}, // 大小写不敏感
		{"image.PNG", true},
		{"image.WEBP", false},
		{"image.tiff", false}, // 不支持的格式
	}

	for _, tt := range tests {
		got := NeedsConversion(tt.path)
		if got != tt.want {
			t.Errorf("NeedsConversion(%q) = %v, want %v", tt.path, got, tt.want)
		}
	}
}
