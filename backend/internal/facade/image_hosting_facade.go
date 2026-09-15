package facade

import (
	"sync/atomic"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

type ImageHostingFacade struct {
	// internal 用原子指针保存：切站（UpdateAppDir）会热替换底层 service，
	// 而 Wails 绑定的是本 facade 对象本身（boot 时绑定一次），只能就地更新指针——
	// 若像旧实现那样替换 AppServices.ImageHosting 字段，Wails 暴露的仍是旧对象、仍指旧站点目录。
	internal atomic.Pointer[service.ImageHostingService]
}

func NewImageHostingFacade(s *service.ImageHostingService) *ImageHostingFacade {
	f := &ImageHostingFacade{}
	f.internal.Store(s)
	return f
}

func (f *ImageHostingFacade) svc() *service.ImageHostingService { return f.internal.Load() }

func (f *ImageHostingFacade) GetSetting() (*domain.ImageHostingSetting, error) {
	return f.svc().GetSetting()
}

func (f *ImageHostingFacade) SaveSetting(setting *domain.ImageHostingSetting) error {
	return f.svc().SaveSetting(setting)
}

func (f *ImageHostingFacade) UploadImage(filePath string) (*domain.ImageHostingFile, error) {
	return f.svc().Upload(filePath)
}

func (f *ImageHostingFacade) ListImages(page int) (*domain.ImageHostingListResponse, error) {
	return f.svc().List(page)
}

func (f *ImageHostingFacade) DeleteImage(hash string) error {
	return f.svc().Delete(hash)
}

func (f *ImageHostingFacade) UploadImagesFromFrontend(files []domain.UploadedFile) ([]string, error) {
	return f.svc().UploadFromFrontend(files)
}
