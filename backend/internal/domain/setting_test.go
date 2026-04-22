package domain

import "testing"

func TestSetting_Clone_IndependentPlatformConfigs(t *testing.T) {
	original := Setting{
		Platform: "github",
		PlatformConfigs: map[string]map[string]any{
			"github": {"username": "user", "repo": "r"},
		},
	}

	cp := original.Clone()

	// 修改 clone 的内层 map，不应影响原始
	cp.PlatformConfigs["github"]["token"] = "secret"
	if _, ok := original.PlatformConfigs["github"]["token"]; ok {
		t.Error("mutating clone should not leak into original inner map")
	}

	// 修改原始的内层 map，不应影响 clone
	original.PlatformConfigs["github"]["username"] = "changed"
	if cp.PlatformConfigs["github"]["username"] == "changed" {
		t.Error("mutating original should not leak into clone")
	}

	// 添加一个新平台到 clone，不应出现在原始
	cp.PlatformConfigs["netlify"] = map[string]any{"site": "abc"}
	if _, ok := original.PlatformConfigs["netlify"]; ok {
		t.Error("adding platform to clone should not leak into original top-level map")
	}
}

func TestSetting_Clone_PreservesOtherFields(t *testing.T) {
	original := Setting{
		Platform:     "sftp",
		ProxyEnabled: true,
		ProxyURL:     "http://127.0.0.1:8080",
	}
	cp := original.Clone()
	if cp.Platform != "sftp" || !cp.ProxyEnabled || cp.ProxyURL != "http://127.0.0.1:8080" {
		t.Errorf("Clone lost scalar fields: %+v", cp)
	}
}

func TestSetting_Clone_NilPlatformConfigsStaysNil(t *testing.T) {
	original := Setting{Platform: "github"}
	cp := original.Clone()
	if cp.PlatformConfigs != nil {
		t.Errorf("nil PlatformConfigs should stay nil in clone, got %v", cp.PlatformConfigs)
	}
}

// 核心回归：InjectCredentials 注入 token 后，只影响传入的 setting，不污染原始 map。
func TestSetting_InjectCredentialsOnClone_DoesNotLeakIntoOriginal(t *testing.T) {
	cache := Setting{
		Platform: "github",
		PlatformConfigs: map[string]map[string]any{
			"github": {"username": "user"},
		},
	}

	// 模拟 repo.GetSetting 返回 Clone
	used := cache.Clone()
	used.InjectCredentials(map[string]string{"github:token": "SECRET_FROM_KEYCHAIN"})

	// used 上能看到 token
	if used.PlatformConfigs["github"]["token"] != "SECRET_FROM_KEYCHAIN" {
		t.Error("token not injected into clone")
	}
	// 原始 cache 不应受影响
	if _, ok := cache.PlatformConfigs["github"]["token"]; ok {
		t.Error("token leaked back into original cache — Clone did not protect inner map")
	}
}
