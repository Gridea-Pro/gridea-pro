package facade

import (
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

type ImageHostingFacade struct {
	internal *service.ImageHostingService
}

func NewImageHostingFacade(s *service.ImageHostingService) *ImageHostingFacade {
	return &ImageHostingFacade{internal: s}
}

func (f *ImageHostingFacade) GetSetting() (*domain.ImageHostingSetting, error) {
	return f.internal.GetSetting()
}

func (f *ImageHostingFacade) SaveSetting(setting *domain.ImageHostingSetting) error {
	return f.internal.SaveSetting(setting)
}

func (f *ImageHostingFacade) UploadImage(filePath string) (*domain.ImageHostingFile, error) {
	return f.internal.Upload(filePath)
}

func (f *ImageHostingFacade) ListImages(page int) (*domain.ImageHostingListResponse, error) {
	return f.internal.List(page)
}

func (f *ImageHostingFacade) DeleteImage(hash string) error {
	return f.internal.Delete(hash)
}

func (f *ImageHostingFacade) UploadImagesFromFrontend(files []domain.UploadedFile) ([]string, error) {
	return f.internal.UploadFromFrontend(files)
}
