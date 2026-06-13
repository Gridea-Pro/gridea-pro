#!/usr/bin/env python3
"""npm audit 门禁：high 及以上漏洞阻断，支持显式豁免。

用法：npm audit --json > audit.json && python3 scripts/check_npm_audit.py audit.json

豁免规则：
- 只豁免评估过"不随安装包分发给用户"的开发期漏洞（dev server、构建工具自身）。
- 每条豁免必须注明理由；上游发布修复版本后应升级依赖并删除对应豁免。
- 运行时依赖（会打进前端 bundle 的库）的漏洞一律不得豁免。
"""

import json
import sys

# GHSA 编号 -> 豁免理由
ALLOWLIST = {
    # rolldown-vite（vite 别名）dev server 漏洞三连：仅影响开发者本机 `wails dev`，
    # 不进用户安装包。rolldown-vite 截至 2026-06 无修复版本，出新版后升级并删除。
    "GHSA-4w7w-66w2-5vf9": "vite dev server 路径穿越，开发期",
    "GHSA-v2wj-q39q-566r": "vite dev server fs.deny 绕过，开发期",
    "GHSA-p9ff-h696-f583": "vite dev server WebSocket 文件读取，开发期",
}

BLOCK_SEVERITIES = {"high", "critical"}


def main(audit_file: str) -> int:
    with open(audit_file) as f:
        data = json.load(f)

    blocking = []
    for name, vuln in data.get("vulnerabilities", {}).items():
        if vuln.get("severity") not in BLOCK_SEVERITIES:
            continue
        # via 含 dict（直接公告）与 str（依赖链转发）两种；只有全部公告都在豁免清单内才放行
        advisories = {
            v["url"].rsplit("/", 1)[-1]
            for v in vuln.get("via", [])
            if isinstance(v, dict) and v.get("url")
        }
        if advisories and advisories <= set(ALLOWLIST):
            for ghsa in sorted(advisories):
                print(f"豁免 {name}: {ghsa}（{ALLOWLIST[ghsa]}）")
            continue
        blocking.append(f"{name} [{vuln.get('severity')}] {sorted(advisories) or vuln.get('via')}")

    if blocking:
        print("\n发现未豁免的 high+ 漏洞，请升级依赖或（仅限开发期漏洞）评估后加入豁免清单：")
        for item in blocking:
            print(f"  - {item}")
        return 1

    print("npm audit 门禁通过")
    return 0


if __name__ == "__main__":
    sys.exit(main(sys.argv[1]))
