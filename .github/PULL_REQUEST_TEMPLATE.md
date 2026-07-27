## 变更说明

<!-- 这个 PR 做了什么、为什么这么做。UI 改动请附前后截图。 -->

## 关联 Issue

<!-- 写法：Closes #123（多个 issue 需逐个带关键词：Closes #1, closes #2, closes #3，
     否则 GitHub 只会自动关闭第一个） -->

## 自查清单

提交后 CI 会自动检查以下各项，任一不过都无法合并；建议先在本地过一遍：

- [ ] 本地 `wails dev` 跑通并自测过改动
- [ ] Go 代码已 `gofmt -w` 格式化，`go vet ./...`、`go test ./backend/...` 通过
- [ ] 前端 `npm run lint` 无 error
- [ ] 新增/修改的 UI 文案已补全 **全部 12 个语言**（`frontend/src/locales/`），
      可运行 `python3 scripts/check_i18n.py frontend/src/locales` 自检；en.json 不要遗留中文
- [ ] 未夹带与本 PR 无关的改动
