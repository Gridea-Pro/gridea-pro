// Package localfileauth 维护 /local-file 端点对「站点目录之外」绝对路径的一次性授权。
//
// 背景：/local-file 既要预览站点目录内的图片（配置里的头像/特色图，路径都在 appDir 下），
// 也要预览用户刚通过系统文件对话框选中的、位于桌面/下载等站点外的图片。前者按目录前缀放行即可；
// 后者无法靠目录判断，否则 webview 里任何 XSS 都能 fetch('/local-file?path=/任意/图片') 外传。
//
// 方案：只有经 OpenImageDialog 等「用户主动选择」入口返回的绝对路径才被 Authorize 登记进内存白名单，
// /local-file 放行站点内路径 + 已授权路径，其余一律拒绝。白名单仅存内存、进程退出即失效。
package localfileauth

import (
	"path/filepath"
	"sync"
	"time"
)

const ttl = 30 * time.Minute

var (
	mu      sync.Mutex
	allowed = map[string]time.Time{} // cleanPath -> 过期时刻
)

// Authorize 登记一个用户主动选择的绝对路径，使其在 TTL 内可经 /local-file 读取。
func Authorize(path string) {
	if path == "" {
		return
	}
	clean := filepath.Clean(path)
	mu.Lock()
	defer mu.Unlock()
	now := time.Now()
	allowed[clean] = now.Add(ttl)
	// 顺手清理过期项，避免长期运行内存无界增长。
	for p, exp := range allowed {
		if exp.Before(now) {
			delete(allowed, p)
		}
	}
}

// IsAuthorized 判断某绝对路径是否在有效授权期内。
func IsAuthorized(path string) bool {
	clean := filepath.Clean(path)
	mu.Lock()
	defer mu.Unlock()
	exp, ok := allowed[clean]
	if !ok {
		return false
	}
	if exp.Before(time.Now()) {
		delete(allowed, clean)
		return false
	}
	return true
}
