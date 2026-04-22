package repository

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"gridea-pro/backend/internal/domain"
)

// 核心回归场景（#39）：调用方在 GetSetting 返回值上 InjectCredentials 注入 token，
// 再次 GetSetting 不应看到这个 token —— 前端侧永远拿不到敏感字段。
func TestSettingRepository_GetSettingReturnsDeepClone(t *testing.T) {
	ctx := context.Background()
	appDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(appDir, "config"), 0o755)

	repo := NewSettingRepository(appDir)
	initial := domain.Setting{
		Platform: "github",
		PlatformConfigs: map[string]map[string]any{
			"github": {"username": "user"},
		},
	}
	if err := repo.SaveSetting(ctx, initial); err != nil {
		t.Fatalf("SaveSetting: %v", err)
	}

	// 第一次 Get：模拟 deploy 路径注入 Keychain 凭证
	first, err := repo.GetSetting(ctx)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	first.InjectCredentials(map[string]string{"github:token": "SECRET"})
	if first.PlatformConfigs["github"]["token"] != "SECRET" {
		t.Fatalf("expected token injected on first copy")
	}

	// 第二次 Get（模拟前端后续读取）—— 不能看到 token
	second, err := repo.GetSetting(ctx)
	if err != nil {
		t.Fatalf("GetSetting 2: %v", err)
	}
	if _, leaked := second.PlatformConfigs["github"]["token"]; leaked {
		t.Errorf("敏感字段反向泄漏到 repo cache：%v", second.PlatformConfigs["github"])
	}
}

// Save 阶段也要走 Clone：调用方保存后继续修改入参 map，不应影响 cache。
func TestSettingRepository_SaveSettingDecouplesFromInput(t *testing.T) {
	ctx := context.Background()
	appDir := t.TempDir()
	_ = os.MkdirAll(filepath.Join(appDir, "config"), 0o755)

	repo := NewSettingRepository(appDir)
	input := domain.Setting{
		Platform: "github",
		PlatformConfigs: map[string]map[string]any{
			"github": {"username": "u1"},
		},
	}
	if err := repo.SaveSetting(ctx, input); err != nil {
		t.Fatalf("SaveSetting: %v", err)
	}

	// 调用方后续修改自己手里的 map（这是 Go map 的常见陷阱）
	input.PlatformConfigs["github"]["username"] = "TAMPERED"

	got, err := repo.GetSetting(ctx)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if got.PlatformConfigs["github"]["username"] != "u1" {
		t.Errorf("repo cache 被输入侧后续修改污染：%v", got.PlatformConfigs["github"])
	}
}
