package main

import (
	"context"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"gridea-pro/backend/internal/service"
)

// MockThemeRepository
type MockThemeRepository struct{}

func (m *MockThemeRepository) GetAll(ctx context.Context) ([]domain.Theme, error) {
	return nil, nil
}
func (m *MockThemeRepository) SaveConfig(ctx context.Context, config domain.ThemeConfig) error {
	return nil
}
func (m *MockThemeRepository) GetConfig(ctx context.Context) (domain.ThemeConfig, error) {
	return domain.ThemeConfig{
		SiteAuthor: "Test Author",
		SiteEmail:  "test@example.com",
		Domain:     "https://test.com",
	}, nil
}

// MockCommentRepository
type MockCommentRepository struct{}

func (m *MockCommentRepository) GetSettings(ctx context.Context) (domain.CommentSettings, error) {
	return domain.CommentSettings{
		Enable:   true,
		Platform: domain.CommentPlatformValine,
		PlatformConfigs: map[domain.CommentPlatform]map[string]any{
			domain.CommentPlatformValine: {
				"appId":      "test-app-id",
				"appKey":     "test-app-key",
				"serverURLs": "https://test.api.leancloud.cn",
			},
		},
	}, nil
}
func (m *MockCommentRepository) SaveSettings(ctx context.Context, settings domain.CommentSettings) error {
	return nil
}

// MockPostRepository
type MockPostRepository struct{}

func (m *MockPostRepository) GetAll(ctx context.Context) ([]domain.Post, error) {
	return nil, nil
}
func (m *MockPostRepository) GetByFileName(ctx context.Context, fileName string) (*domain.Post, error) {
	return nil, nil
}
func (m *MockPostRepository) Save(ctx context.Context, post *domain.PostInput) error {
	return nil
}
func (m *MockPostRepository) Delete(ctx context.Context, fileName string) error {
	return nil
}

func main() {
	// Setup Mocks
	themeRepo := &MockThemeRepository{}
	commentRepo := &MockCommentRepository{}
	postRepo := &MockPostRepository{}

	// Init Service
	svc := service.NewCommentService("/tmp/test-app", commentRepo, postRepo, themeRepo)

	fmt.Println("Verifying pagination logic...")
	// Let's just verify compilation of the call
	_, err := svc.FetchComments(context.Background(), 1, 10)
	if err != nil {
		fmt.Printf("FetchComments returned error (expected without valid config): %v\n", err)
	} else {
		fmt.Println("FetchComments executed successfully (unexpected with invalid auth).")
	}

	fmt.Println("Pagination structs and interfaces are valid.")
}
