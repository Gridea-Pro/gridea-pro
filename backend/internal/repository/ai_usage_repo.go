package repository

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"path/filepath"
	"sync"

	"gridea-pro/backend/internal/domain"
)

// 内置 HMAC secret，写死在二进制里
// 普通用户无法逆向得到该值，因此无法手动伪造合法 sig
const aiUsageSigSecret = "gridea-pro-ai-usage-v1-7f3a9c2e8b1d4f6a"

// aiUsageRepository 内置模型调用计数器
//
// 存储位置：应用级配置目录 ai_usage.json
//   - macOS:   ~/Library/Application Support/Gridea Pro/ai_usage.json
//   - Linux:   ~/.config/Gridea Pro/ai_usage.json
//   - Windows: %AppData%/Gridea Pro/ai_usage.json
//
// 放在应用级目录而非站点目录的原因：
//  1. 配额是设备级语义，不应该因切换站点而重置
//  2. 不应跟随站点同步到云端，避免被多设备绕过
type aiUsageRepository struct {
	mu        sync.RWMutex
	configDir string
	cache     *domain.AIUsage
	loaded    bool
}

func NewAIUsageRepository(appConfigDir string) domain.AIUsageRepository {
	return &aiUsageRepository{configDir: appConfigDir}
}

func (r *aiUsageRepository) filePath() string {
	return filepath.Join(r.configDir, "ai_usage.json")
}

// computeUsageSig 根据字段计算 HMAC-SHA256 签名
func computeUsageSig(u domain.AIUsage) string {
	payload := fmt.Sprintf("v1|%s|%d|%s|%d", u.Date, u.DailyCount, u.Minute, u.MinuteCount)
	mac := hmac.New(sha256.New, []byte(aiUsageSigSecret))
	mac.Write([]byte(payload))
	return hex.EncodeToString(mac.Sum(nil))
}

// verifyUsageSig 校验 sig 是否合法
func verifyUsageSig(u domain.AIUsage) bool {
	if u.Sig == "" {
		return false
	}
	expected := computeUsageSig(u)
	return hmac.Equal([]byte(u.Sig), []byte(expected))
}

func (r *aiUsageRepository) loadIfNeeded() error {
	r.mu.RLock()
	if r.loaded {
		r.mu.RUnlock()
		return nil
	}
	r.mu.RUnlock()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.loaded {
		return nil
	}

	var usage domain.AIUsage
	if err := LoadJSONFile(r.filePath(), &usage); err != nil {
		// 只有「文件不存在」才视为新设备、计数从 0 开始。
		// IO/权限/JSON 损坏等错误不能重置为 0——否则杀软锁文件或磁盘异常就成了绕过配额的路径。
		// 这类错误 fail-closed：不缓存、不置 loaded，向上抛错让 reserveBuiltInQuota 拒绝本次调用并可重试。
		if errors.Is(err, fs.ErrNotExist) {
			r.cache = &domain.AIUsage{}
			r.loaded = true
			return nil
		}
		return err
	}

	// 文件存在但 sig 校验失败 → 用户篡改过，静默重置（反作弊语义，与 IO 错误区别对待）
	if !verifyUsageSig(usage) {
		r.cache = &domain.AIUsage{}
		r.loaded = true
		return nil
	}

	r.cache = &usage
	r.loaded = true
	return nil
}

func (r *aiUsageRepository) GetAIUsage(ctx context.Context) (domain.AIUsage, error) {
	if err := r.loadIfNeeded(); err != nil {
		return domain.AIUsage{}, err
	}

	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.cache == nil {
		return domain.AIUsage{}, nil
	}
	return *r.cache, nil
}

func (r *aiUsageRepository) SaveAIUsage(ctx context.Context, usage domain.AIUsage) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// 写入前重新计算 sig，覆盖任何旧值
	usage.Sig = computeUsageSig(usage)

	if err := SaveJSONFile(r.filePath(), usage); err != nil {
		return err
	}
	r.cache = &usage
	r.loaded = true
	return nil
}
