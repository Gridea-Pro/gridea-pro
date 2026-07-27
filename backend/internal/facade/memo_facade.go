package facade

import (
	"context"
	"sync/atomic"
	"time"

	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// parseMemoCreatedAt 解析前端传入的发布时间字符串。
// 空串或无法识别的格式返回零值，CreateMemo 会据此回退到当前时间。
func parseMemoCreatedAt(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	for _, layout := range []string{time.RFC3339, "2006-01-02 15:04:05"} {
		if t, err := time.ParseInLocation(layout, s, time.Local); err == nil {
			return t
		}
	}
	return time.Time{}
}

// MemoFacade wraps MemoService
type MemoFacade struct {
	// internal 用原子读写保存：切站（UpdateAppDir）会热替换它，
	// 而 Wails 可能并发调用本 facade 的方法，原子读写避免"读到替换一半的状态"的 data race。
	internal atomic.Pointer[service.MemoService]
}

func NewMemoFacade(s *service.MemoService) *MemoFacade {
	f := &MemoFacade{}
	f.internal.Store(s)
	return f
}

func (f *MemoFacade) svc() *service.MemoService { return f.internal.Load() }

// dashboard 用给定的 svc 快照组装 memos + stats，供各写方法复用，
// 保证"写"与随后的"读列表"落在同一站点快照，避免中途切站串号。
func (f *MemoFacade) dashboard(ctx context.Context, svc *service.MemoService) (*domain.MemoDashboardDTO, error) {
	memos, err := svc.LoadMemos(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := svc.GetMemoStats(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.MemoDashboardDTO{
		Memos: memos,
		Stats: *stats,
	}, nil
}

// LoadMemosFromFrontend wraps LoadMemos and returns memos and stats
func (f *MemoFacade) LoadMemosFromFrontend() (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	return f.dashboard(ctx, f.svc())
}

// SaveMemoFromFrontend saves a new memo and returns updated list.
// createdAt 为空串时按当前时间发布；非空时按指定时间发布。
func (f *MemoFacade) SaveMemoFromFrontend(content string, createdAt string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if _, err := svc.CreateMemo(ctx, content, parseMemoCreatedAt(createdAt)); err != nil {
		return nil, err
	}
	return f.dashboard(ctx, svc)
}

// UpdateMemoFromFrontend updates a memo and returns updated list
func (f *MemoFacade) UpdateMemoFromFrontend(memo domain.Memo) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if err := svc.UpdateMemo(ctx, memo); err != nil {
		return nil, err
	}
	return f.dashboard(ctx, svc)
}

// DeleteMemoFromFrontend deletes a memo and returns updated list
func (f *MemoFacade) DeleteMemoFromFrontend(id string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if err := svc.DeleteMemo(ctx, id); err != nil {
		return nil, err
	}
	return f.dashboard(ctx, svc)
}

// RenameMemoTagFromFrontend renames a tag and returns updated list
func (f *MemoFacade) RenameMemoTagFromFrontend(oldName, newName string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if err := svc.RenameTag(ctx, oldName, newName); err != nil {
		return nil, err
	}
	return f.dashboard(ctx, svc)
}

// DeleteMemoTagFromFrontend deletes a tag and returns updated list
func (f *MemoFacade) DeleteMemoTagFromFrontend(tagName string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	svc := f.svc()
	if err := svc.DeleteTag(ctx, tagName); err != nil {
		return nil, err
	}
	return f.dashboard(ctx, svc)
}

// LoadMemos (Deprecated: use Service directly or LoadMemosFromFrontend)
func (f *MemoFacade) LoadMemos() ([]domain.Memo, error) {
	return f.svc().LoadMemos(context.TODO())
}

// SaveMemos (Deprecated: use Service directly)
func (f *MemoFacade) SaveMemos(memos []domain.Memo) error {
	return f.svc().SaveMemos(context.TODO(), memos)
}

// GetMemoStats (Deprecated: use Service directly)
func (f *MemoFacade) GetMemoStats() (*domain.MemoStats, error) {
	return f.svc().GetMemoStats(context.TODO())
}

// RegisterEvents 注册闪念相关事件监听器
// No longer registers backend-side event listeners for CRUD.
// Frontend should call exported methods directly.
func (f *MemoFacade) RegisterEvents(ctx context.Context) {
	// Intentionally empty.
	// Previous logic has been migrated to synchronous Wails methods.
}
