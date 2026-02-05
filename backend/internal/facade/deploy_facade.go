package facade

import (
	"gridea-pro/backend/internal/service"
)

// DeployFacade wraps DeployService
type DeployFacade struct {
	internal *service.DeployService
}

func NewDeployFacade(s *service.DeployService) *DeployFacade {
	return &DeployFacade{internal: s}
}

func (f *DeployFacade) DeployToGit() error {
	return f.internal.DeployToGit()
}
