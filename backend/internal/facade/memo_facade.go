package facade

import (
	"context"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// MemoFacade wraps MemoService
type MemoFacade struct {
	internal *service.MemoService
}

func NewMemoFacade(s *service.MemoService) *MemoFacade {
	return &MemoFacade{internal: s}
}

// LoadMemosFromFrontend wraps LoadMemos and returns memos and stats
func (f *MemoFacade) LoadMemosFromFrontend() (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	memos, err := f.internal.LoadMemos(ctx)
	if err != nil {
		return nil, err
	}
	stats, err := f.internal.GetMemoStats(ctx)
	if err != nil {
		return nil, err
	}
	return &domain.MemoDashboardDTO{
		Memos: memos,
		Stats: *stats,
	}, nil
}

// SaveMemoFromFrontend saves a new memo and returns updated list
func (f *MemoFacade) SaveMemoFromFrontend(content string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	_, err := f.internal.CreateMemo(ctx, content)
	if err != nil {
		return nil, err
	}
	return f.LoadMemosFromFrontend()
}

// UpdateMemoFromFrontend updates a memo and returns updated list
func (f *MemoFacade) UpdateMemoFromFrontend(memo domain.Memo) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	err := f.internal.UpdateMemo(ctx, memo)
	if err != nil {
		return nil, err
	}
	return f.LoadMemosFromFrontend()
}

// DeleteMemoFromFrontend deletes a memo and returns updated list
func (f *MemoFacade) DeleteMemoFromFrontend(id string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	err := f.internal.DeleteMemo(ctx, id)
	if err != nil {
		return nil, err
	}
	return f.LoadMemosFromFrontend()
}

// RenameMemoTagFromFrontend renames a tag and returns updated list
func (f *MemoFacade) RenameMemoTagFromFrontend(oldName, newName string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	if err := f.internal.RenameTag(ctx, oldName, newName); err != nil {
		return nil, err
	}
	return f.LoadMemosFromFrontend()
}

// DeleteMemoTagFromFrontend deletes a tag and returns updated list
func (f *MemoFacade) DeleteMemoTagFromFrontend(tagName string) (*domain.MemoDashboardDTO, error) {
	ctx := WailsContext
	if ctx == nil {
		ctx = context.TODO()
	}
	if err := f.internal.DeleteTag(ctx, tagName); err != nil {
		return nil, err
	}
	return f.LoadMemosFromFrontend()
}

// LoadMemos (Deprecated: use Service directly or LoadMemosFromFrontend)
func (f *MemoFacade) LoadMemos() ([]domain.Memo, error) {
	return f.internal.LoadMemos(context.TODO())
}

// SaveMemos (Deprecated: use Service directly)
func (f *MemoFacade) SaveMemos(memos []domain.Memo) error {
	return f.internal.SaveMemos(context.TODO(), memos)
}

// GetMemoStats (Deprecated: use Service directly)
func (f *MemoFacade) GetMemoStats() (*domain.MemoStats, error) {
	return f.internal.GetMemoStats(context.TODO())
}

// RegisterEvents 注册闪念相关事件监听器
// No longer registers backend-side event listeners for CRUD.
// Frontend should call exported methods directly.
func (f *MemoFacade) RegisterEvents(ctx context.Context) {
	// Intentionally empty.
	// Previous logic has been migrated to synchronous Wails methods.
}
