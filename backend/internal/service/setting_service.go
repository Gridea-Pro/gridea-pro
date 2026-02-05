package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"io"
	"os"
	"path/filepath"
)

type SettingService struct {
	repo   domain.SettingRepository
	appDir string
}

func NewSettingService(appDir string, repo domain.SettingRepository) *SettingService {
	return &SettingService{
		appDir: appDir,
		repo:   repo,
	}
}

func (s *SettingService) SaveAvatar(sourcePath string) error {
	destPath := filepath.Join(s.appDir, "images", "avatar.png")
	return s.copyFile(sourcePath, destPath)
}

func (s *SettingService) SaveFavicon(sourcePath string) error {
	destPath := filepath.Join(s.appDir, "favicon.ico")
	return s.copyFile(sourcePath, destPath)
}

func (s *SettingService) copyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	// Ensure destination directory exists
	if err := os.MkdirAll(filepath.Dir(dst), 0755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	if _, err := io.Copy(destination, source); err != nil {
		return err
	}
	return nil
}

func (s *SettingService) GetSetting(ctx context.Context) (domain.Setting, error) {
	return s.repo.GetSetting(ctx)
}

func (s *SettingService) SaveSetting(ctx context.Context, setting domain.Setting) error {
	return s.repo.SaveSetting(ctx, setting)
}
