package engine

import (
	"encoding/json"
	"fmt"
	"gridea-pro/backend/internal/domain"
	"log/slog"
	"path/filepath"
	"time"
)

// PwaGenerator 生成 PWA 相关文件（manifest.json、sw.js、offline.html）
type PwaGenerator struct {
	appDir   string
	logger   *slog.Logger
	manifest *RenderManifest
}

func NewPwaGenerator(appDir string) *PwaGenerator {
	return &PwaGenerator{
		appDir: appDir,
		logger: slog.Default(),
	}
}

// SetManifest 设置渲染产物跟踪器
func (g *PwaGenerator) SetManifest(m *RenderManifest) {
	g.manifest = m
}

type manifestIcon struct {
	Src     string `json:"src"`
	Sizes   string `json:"sizes"`
	Type    string `json:"type"`
	Purpose string `json:"purpose,omitempty"`
}

type manifest struct {
	ID              string         `json:"id"`
	Name            string         `json:"name"`
	ShortName       string         `json:"short_name"`
	Description     string         `json:"description,omitempty"`
	Lang            string         `json:"lang,omitempty"`
	StartURL        string         `json:"start_url"`
	Scope           string         `json:"scope"`
	Display         string         `json:"display"`
	DisplayOverride []string       `json:"display_override,omitempty"`
	Orientation     string         `json:"orientation,omitempty"`
	BackgroundColor string         `json:"background_color"`
	ThemeColor      string         `json:"theme_color"`
	Categories      []string       `json:"categories,omitempty"`
	Icons           []manifestIcon `json:"icons,omitempty"`
}

// RenderManifest 生成 manifest.json
// language 从主题配置的站点语言获取
func (g *PwaGenerator) RenderManifest(buildDir string, setting *domain.PwaSetting, siteName, siteLanguage string) error {
	themeColor := setting.ThemeColor
	if themeColor == "" {
		themeColor = "#ffffff"
	}
	bgColor := setting.BackgroundColor
	if bgColor == "" {
		bgColor = "#ffffff"
	}
	appName := setting.AppName
	if appName == "" {
		appName = siteName
	}
	shortName := setting.ShortName
	if shortName == "" {
		shortName = appName
	}
	orientation := setting.Orientation
	if orientation == "" {
		orientation = "any"
	}

	m := manifest{
		ID:              "/",
		Name:            appName,
		ShortName:       shortName,
		Description:     setting.Description,
		Lang:            siteLanguage,
		StartURL:        "/",
		Scope:           "/",
		Display:         "standalone",
		DisplayOverride: []string{"standalone", "minimal-ui"},
		Orientation:     orientation,
		BackgroundColor: bgColor,
		ThemeColor:      themeColor,
		Categories:      []string{"blog"},
	}

	// 图标处理：优先使用自定义 PWA 图标，否则从头像生成
	if setting.CustomIcon {
		// 从自定义 PWA 图标生成各尺寸
		bgParsed := parseHexColor(bgColor)
		icons, err := generateAllPwaIconsFromSource(g.appDir, buildDir, "pwa-icon.png", bgParsed)
		if err != nil {
			g.logger.Warn(fmt.Sprintf("从自定义图标生成 PWA 图标失败，尝试使用头像: %v", err))
			icons, err = generateAllPwaIcons(g.appDir, buildDir, bgParsed)
		}
		if err != nil {
			g.logger.Warn(fmt.Sprintf("生成 PWA 图标失败: %v", err))
		} else {
			g.appendIcons(&m, icons)
		}
	} else {
		// 从头像自动生成完整图标集（圆角 + maskable）
		bgParsed := parseHexColor(bgColor)
		icons, err := generateAllPwaIcons(g.appDir, buildDir, bgParsed)
		if err != nil {
			g.logger.Warn(fmt.Sprintf("从头像生成 PWA 图标失败: %v", err))
		} else {
			g.appendIcons(&m, icons)
		}
	}

	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return fmt.Errorf("生成 manifest.json 失败: %w", err)
	}

	if err := g.manifest.WriteFile(filepath.Join(buildDir, "manifest.json"), data, 0644); err != nil {
		return fmt.Errorf("写入 manifest.json 失败: %w", err)
	}

	g.logger.Info("✅ 已生成 manifest.json")
	return nil
}

// appendIcons 向 manifest 添加完整图标集
func (g *PwaGenerator) appendIcons(m *manifest, icons *pwaIconSet) {
	// 圆角图标 (purpose: any)
	m.Icons = append(m.Icons,
		manifestIcon{Src: icons.Icon192, Sizes: "192x192", Type: "image/png", Purpose: "any"},
		manifestIcon{Src: icons.Icon512, Sizes: "512x512", Type: "image/png", Purpose: "any"},
	)
	// Maskable 图标 (purpose: maskable) — Android 自适应图标
	m.Icons = append(m.Icons,
		manifestIcon{Src: icons.Maskable192, Sizes: "192x192", Type: "image/png", Purpose: "maskable"},
		manifestIcon{Src: icons.Maskable512, Sizes: "512x512", Type: "image/png", Purpose: "maskable"},
	)
	g.logger.Info("✅ 已生成 PWA 图标集（圆角 + maskable）")
}

// RenderServiceWorker 生成 sw.js（含离线回退和更新检测）
func (g *PwaGenerator) RenderServiceWorker(buildDir string) error {
	cacheVersion := fmt.Sprintf("gridea-v%d", time.Now().Unix())

	swContent := fmt.Sprintf(`// Gridea Pro PWA Service Worker
var CACHE_NAME = '%s';
var OFFLINE_URL = '/offline.html';

// 安装：预缓存离线页面
self.addEventListener('install', function(event) {
  event.waitUntil(
    caches.open(CACHE_NAME).then(function(cache) {
      return cache.add(OFFLINE_URL);
    }).then(function() {
      return self.skipWaiting();
    })
  );
});

// 激活：清理旧缓存并接管所有客户端
self.addEventListener('activate', function(event) {
  event.waitUntil(
    caches.keys().then(function(names) {
      return Promise.all(
        names.filter(function(name) {
          return name !== CACHE_NAME;
        }).map(function(name) {
          return caches.delete(name);
        })
      );
    }).then(function() {
      return self.clients.claim();
    })
  );
});

// 请求拦截
self.addEventListener('fetch', function(event) {
  var request = event.request;

  if (request.method !== 'GET') return;
  if (!request.url.startsWith(self.location.origin)) return;

  // 导航请求（HTML）：network-first，失败时返回离线页面
  if (request.mode === 'navigate') {
    event.respondWith(
      fetch(request).then(function(response) {
        if (response.ok) {
          var clone = response.clone();
          caches.open(CACHE_NAME).then(function(cache) {
            cache.put(request, clone);
          });
        }
        return response;
      }).catch(function() {
        return caches.match(request).then(function(cached) {
          return cached || caches.match(OFFLINE_URL);
        });
      })
    );
    return;
  }

  // 静态资源：cache-first
  if (isStaticAsset(request.url)) {
    event.respondWith(
      caches.match(request).then(function(cached) {
        if (cached) return cached;
        return fetch(request).then(function(response) {
          if (response.ok) {
            var clone = response.clone();
            caches.open(CACHE_NAME).then(function(cache) {
              cache.put(request, clone);
            });
          }
          return response;
        });
      })
    );
    return;
  }
});

function isStaticAsset(url) {
  return /\.(css|js|png|jpg|jpeg|gif|svg|webp|ico|woff|woff2|ttf|eot|json)(\?.*)?$/i.test(url);
}
`, cacheVersion)

	if err := g.manifest.WriteFile(filepath.Join(buildDir, "sw.js"), []byte(swContent), 0644); err != nil {
		return fmt.Errorf("写入 sw.js 失败: %w", err)
	}

	g.logger.Info("✅ 已生成 sw.js")
	return nil
}

// RenderOfflinePage 生成离线回退页面
func (g *PwaGenerator) RenderOfflinePage(buildDir string, siteName, themeColor string) error {
	if themeColor == "" {
		themeColor = "#6366f1"
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="zh">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width,initial-scale=1">
<title>%s - 离线</title>
<style>
*{margin:0;padding:0;box-sizing:border-box}
body{font-family:-apple-system,BlinkMacSystemFont,"Segoe UI",Roboto,sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;background:#f9fafb;color:#374151;text-align:center;padding:2rem}
.container{max-width:400px}
.icon{font-size:4rem;margin-bottom:1.5rem;opacity:.6}
h1{font-size:1.5rem;font-weight:600;margin-bottom:.75rem}
p{color:#6b7280;line-height:1.6;margin-bottom:2rem}
button{background:%s;color:#fff;border:none;padding:.75rem 2rem;border-radius:2rem;font-size:.9rem;cursor:pointer;transition:opacity .2s}
button:hover{opacity:.85}
</style>
</head>
<body>
<div class="container">
<div class="icon">📡</div>
<h1>当前无法连接网络</h1>
<p>请检查你的网络连接后重试。已访问过的页面可能仍然可用。</p>
<button onclick="location.reload()">重新加载</button>
</div>
</body>
</html>`, escapeAttr(siteName), escapeAttr(themeColor))

	if err := g.manifest.WriteFile(filepath.Join(buildDir, "offline.html"), []byte(html), 0644); err != nil {
		return fmt.Errorf("写入 offline.html 失败: %w", err)
	}

	g.logger.Info("✅ 已生成 offline.html")
	return nil
}
