package service

// oauthResultHTML 生成 OAuth 回调结果页面 HTML
// lang: 客户端语言，zh/zh-CN/zh-cn 显示中文，其他显示英文
func oauthResultHTML(success bool, providerID, lang, errMsg, username, avatarURL string) string {
	platformName := platformDisplayNames[providerID]
	if platformName == "" {
		platformName = providerID
	}

	t := oauthTextsEN
	if isZhLang(lang) {
		t = oauthTextsZH
	}

	const pageStyle = `*{box-sizing:border-box;margin:0;padding:0}
body{font-family:-apple-system,BlinkMacSystemFont,'Segoe UI','Helvetica Neue',sans-serif;display:flex;align-items:center;justify-content:center;min-height:100vh;background:#000}
.card{text-align:center;padding:48px 56px;background:rgba(255,255,255,.06);border:1px solid rgba(255,255,255,.08);border-radius:24px;backdrop-filter:blur(20px);max-width:420px;width:100%}
.brand{margin-bottom:32px}
.brand img{width:72px;height:72px;border-radius:14px;margin-bottom:10px}
.brand-name{font-size:20px;font-weight:600;color:rgba(255,255,255,.9);letter-spacing:-.2px}
.status-icon{width:52px;height:52px;border-radius:50%;display:flex;align-items:center;justify-content:center;margin:0 auto 12px}
.status-icon.ok{background:linear-gradient(135deg,#34c759,#30d158)}
.status-icon.fail{background:linear-gradient(135deg,#ff3b30,#ff453a)}
.status-icon svg{width:26px;height:26px;color:#fff}
.title{font-size:16px;font-weight:700;color:#34c759;margin-bottom:16px}
.title.fail{color:#ff453a}
.user-row{display:flex;align-items:center;justify-content:center;gap:10px;margin-bottom:8px}
.user-row img{width:24px;height:24px;border-radius:50%;border:1px solid rgba(255,255,255,.3)}
.user-row span{font-size:14px;font-weight:600;color:rgba(255,255,255,.85)}
.hint{font-size:13px;color:rgba(255,255,255,.4);line-height:1.6}
.err{font-size:13px;color:rgba(255,255,255,.5);line-height:1.6;word-break:break-all}
.divider{height:1px;background:rgba(255,255,255,.06);margin:28px 0 20px}
.footer{font-size:11px;color:rgba(255,255,255,.2)}`

	const logoURL = "https://www.gridea.pro/gridea-pro.png"
	const faviconLink = `<link rel="icon" href="https://www.gridea.pro/favicon.ico">`

	checkSVG := `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M5 13l4 4L19 7"/></svg>`
	crossSVG := `<svg fill="none" viewBox="0 0 24 24" stroke="currentColor" stroke-width="2.5"><path stroke-linecap="round" stroke-linejoin="round" d="M6 18L18 6M6 6l12 12"/></svg>`

	brandHTML := `<div class="brand"><img src="` + logoURL + `" alt="Gridea Pro" onerror="this.style.display='none'" /><div class="brand-name">Gridea Pro</div></div>`
	footerHTML := `<div class="divider"></div><div class="footer">Gridea Pro · ` + t.slogan + `</div>`

	if success {
		name := username
		if name == "" {
			name = t.defaultUser
		}
		userHTML := ""
		if avatarURL != "" {
			userHTML = `<div class="user-row"><img src="` + avatarURL + `" alt="" /><span>` + name + ` ` + t.connected + `</span></div>`
		} else {
			userHTML = `<div class="user-row"><span>` + name + ` ` + t.connected + `</span></div>`
		}
		return `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Gridea Pro - ` + platformName + ` ` + t.authSuccess + `</title>` + faviconLink + `
<style>` + pageStyle + `</style></head>
<body><div class="card">
` + brandHTML + `
<div class="status-icon ok">` + checkSVG + `</div>
<div class="title">` + platformName + ` ` + t.authSuccess + `</div>
` + userHTML + `
<div class="hint">` + t.hint + `</div>
` + footerHTML + `
</div></body></html>`
	}

	return `<!DOCTYPE html><html><head><meta charset="UTF-8"><title>Gridea Pro - ` + platformName + ` ` + t.authFailed + `</title>` + faviconLink + `
<style>` + pageStyle + `</style></head>
<body><div class="card">
` + brandHTML + `
<div class="status-icon fail">` + crossSVG + `</div>
<div class="title fail">` + platformName + ` ` + t.authFailed + `</div>
<div class="err">` + errMsg + `</div>
` + footerHTML + `
</div></body></html>`
}

// ─── i18n ────────────────────────────────────────────────────────────────

type oauthTexts struct {
	authSuccess string
	authFailed  string
	connected   string
	defaultUser string
	hint        string
	slogan      string
}

var oauthTextsZH = oauthTexts{
	authSuccess: "授权成功",
	authFailed:  "授权失败",
	connected:   "已连接",
	defaultUser: "账号",
	hint:        "请返回 Gridea Pro 查看，可以关闭此标签页",
	slogan:      "下一代桌面静态博客写作客户端",
}

var oauthTextsEN = oauthTexts{
	authSuccess: "Authorization Successful",
	authFailed:  "Authorization Failed",
	connected:   "connected",
	defaultUser: "Account",
	hint:        "Please return to Gridea Pro. You can close this tab.",
	slogan:      "A modern, AI-powered static blog client.",
}

func isZhLang(lang string) bool {
	if len(lang) >= 2 && (lang[:2] == "zh" || lang[:2] == "Zh" || lang[:2] == "ZH") {
		return true
	}
	return false
}

var platformDisplayNames = map[string]string{
	"github":  "GitHub",
	"gitee":   "Gitee",
	"netlify": "Netlify",
	"vercel":  "Vercel",
	"coding":  "Coding",
	"sftp":    "SFTP",
}
