package facade

import (
	"context"
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
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.internal.DeployToGit(ctx)
}
