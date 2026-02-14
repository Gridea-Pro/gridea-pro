package service

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type DeployService struct {
	settingRepo domain.SettingRepository
	appDir      string
	mu          sync.Mutex
	isDeploying bool
}

func NewDeployService(settingRepo domain.SettingRepository, appDir string) *DeployService {
	return &DeployService{
		settingRepo: settingRepo,
		appDir:      appDir,
	}
}

func (s *DeployService) DeployToGit(ctx context.Context) error {
	s.mu.Lock()
	if s.isDeploying {
		s.mu.Unlock()
		return fmt.Errorf("deployment is already in progress")
	}
	s.isDeploying = true
	s.mu.Unlock()

	// Ensure we reset the flag when done
	defer func() {
		s.mu.Lock()
		s.isDeploying = false
		s.mu.Unlock()
	}()

	s.log(ctx, "Starting deployment check...")

	// 1. Get Settings safely
	setting, err := s.settingRepo.GetSetting(ctx)
	if err != nil {
		s.log(ctx, fmt.Sprintf("Failed to load settings: %v", err))
		return err
	}

	s.log(ctx, fmt.Sprintf("Deploying to domain: %s", setting.Domain))

	// Mock deployment logic
	// In real implementation:
	// 1. Build Static Site
	// 2. Commit & Push to Remote Repo
	s.log(ctx, "Executing git commands (Simulation)...")

	// Simulate work if needed, or just return success
	s.log(ctx, "Deployment successful!")

	return nil
}

// log sends a message to the frontend safely
func (s *DeployService) log(ctx context.Context, msg string) {
	if ctx != nil {
		runtime.EventsEmit(ctx, "deploy-log", msg)
	}
}
