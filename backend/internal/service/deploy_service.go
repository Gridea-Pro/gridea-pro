package service

import (
	"context"
	"gridea-pro/backend/internal/domain"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DeployService struct {
	settingRepo domain.SettingRepository
	appDir      string
	ctx         context.Context // For emitting logs? Or pass ctx in methods.
}

func NewDeployService(settingRepo domain.SettingRepository, appDir string) *DeployService {
	return &DeployService{
		settingRepo: settingRepo,
		appDir:      appDir,
	}
}

func (s *DeployService) SetContext(ctx context.Context) {
	s.ctx = ctx
}

func (s *DeployService) DeployToGit() error {
	// Simplify dependency: fetch setting
	setting, err := s.settingRepo.GetSetting(context.Background())
	if err != nil {
		return err
	}

	// Mock deployment logic
	// In real app, this executes git commands in s.appDir/output
	s.log("Starting deployment to " + setting.Domain)
	s.log("Deploy success (Mock)")

	return nil
}

func (s *DeployService) log(msg string) {
	if s.ctx != nil {
		runtime.EventsEmit(s.ctx, "deploy-log", msg)
	}
}
