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
	return f.internal.DeployToRemote(ctx)
}

// CancelDeploy 对外暴露给前端调用；正在进行的部署会被 ctx 取消（见 #42）。
// 空闲时 no-op。
func (f *DeployFacade) CancelDeploy() {
	f.internal.CancelDeploy()
}

// IsDeploying 返回当前是否有部署在进行，前端按钮状态同步时使用。
func (f *DeployFacade) IsDeploying() bool {
	return f.internal.IsDeploying()
}
