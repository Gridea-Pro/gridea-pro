#!/usr/bin/env python3
"""i18n 完整性检查（CI 用）。

规则：
1. 以 zh-CN.json 为基准，其余每个语言文件的 key 集合必须完整覆盖基准（缺 key → 失败）。
2. 语言文件多出基准没有的 key → 仅警告（不阻断），便于先行翻译。
3. en.json 的值中不得出现 CJK 字符（防止中文文案误留在英文里）→ 失败。

用法：python3 scripts/check_i18n.py [locales_dir]
"""

import json
import re
import sys
from pathlib import Path

CJK = re.compile(r"[一-鿿㐀-䶿]")
BASE = "zh-CN.json"


def flatten(obj, prefix=""):
    out = {}
    for k, v in obj.items():
        key = f"{prefix}.{k}" if prefix else k
        if isinstance(v, dict):
            out.update(flatten(v, key))
        else:
            out[key] = v
    return out


def main():
    locales_dir = Path(sys.argv[1] if len(sys.argv) > 1 else "frontend/src/locales")
    base_path = locales_dir / BASE
    if not base_path.exists():
        print(f"❌ 基准文件不存在: {base_path}")
        return 1

    base = flatten(json.loads(base_path.read_text(encoding="utf-8")))
    failed = False

    for f in sorted(locales_dir.glob("*.json")):
        if f.name == BASE:
            continue
        try:
            loc = flatten(json.loads(f.read_text(encoding="utf-8")))
        except json.JSONDecodeError as e:
            print(f"❌ {f.name}: JSON 解析失败 — {e}")
            failed = True
            continue

        missing = [k for k in base if k not in loc]
        extra = [k for k in loc if k not in base]
        if missing:
            failed = True
            print(f"❌ {f.name}: 缺少 {len(missing)} 个 key（相对 {BASE}）:")
            for k in missing:
                print(f"     - {k}")
        if extra:
            print(f"⚠️  {f.name}: 有 {len(extra)} 个基准之外的 key（不阻断）: {', '.join(extra[:10])}{' ...' if len(extra) > 10 else ''}")

        if f.name == "en.json":
            cjk_hits = [(k, v) for k, v in loc.items() if isinstance(v, str) and CJK.search(v)]
            if cjk_hits:
                failed = True
                print(f"❌ en.json: {len(cjk_hits)} 个值含 CJK 字符（英文文案疑似未翻译）:")
                for k, v in cjk_hits:
                    print(f"     - {k} = {v!r}")

        if not missing and not extra and (f.name != "en.json" or not failed):
            pass

    if failed:
        print("\ni18n 检查未通过：请补全缺失翻译（参照 zh-CN.json），勿在 en.json 留中文。")
        return 1
    print("✅ i18n 检查通过：所有语言 key 完整，en.json 无 CJK 残留。")
    return 0


if __name__ == "__main__":
    sys.exit(main())
