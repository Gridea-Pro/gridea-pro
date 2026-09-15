package facade

import (
	"context"
	"sync/atomic"

	"gridea-pro/backend/internal/service"
)

// DeployFacade wraps DeployService
type DeployFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.DeployService]
}

func NewDeployFacade(s *service.DeployService) *DeployFacade {
	f := &DeployFacade{}
	f.internal.Store(s)
	return f
}

func (f *DeployFacade) svc() *service.DeployService { return f.internal.Load() }

func (f *DeployFacade) DeployToGit() error {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.svc().DeployToRemote(ctx)
}

// CancelDeploy 对外暴露给前端调用；正在进行的部署会被 ctx 取消（见 #42）。
// 空闲时 no-op。
func (f *DeployFacade) CancelDeploy() {
	f.svc().CancelDeploy()
}

// IsDeploying 返回当前是否有部署在进行，前端按钮状态同步时使用。
func (f *DeployFacade) IsDeploying() bool {
	return f.svc().IsDeploying()
}
