/**
 * 守护「排版比例」在两处的同步。
 *
 * index.html 的防闪脚本必须内联一份字阶比例——它在首帧前同步执行，无法 import 模块，
 * 而不内联的话调过字号的用户会先看到一帧 16px 默认排版再跳到自己的设置。
 * 代价是同一份比例存在两地，漂移了不会有任何报错，只会表现为「首屏闪一下字号」
 * 这种极难归因的现象。本脚本把这件事变成一条会失败的检查。
 *
 * 用法：npx tsx scripts/check-typography-sync.mts
 */
import fs from 'node:fs'
import { HEADING_RATIOS, TITLE_RATIO, PROSE_SIZES, PROSE_SIZE_DEFAULT, PROSE_FONTS } from '../src/helpers/typography.ts'

const html = fs.readFileSync(new URL('../index.html', import.meta.url), 'utf-8')
const fails: string[] = []

function nums(re: RegExp, label: string): number[] {
  const m = re.exec(html)
  if (!m) {
    fails.push(`index.html 里找不到 ${label}（防闪脚本是否被改动？）`)
    return []
  }
  return m[1].split(',').map((s) => Number(s.trim()))
}

const eq = (a: readonly number[], b: readonly number[]) =>
  a.length === b.length && a.every((v, i) => v === b[i])

// 1. 标题字阶比例
const ratios = nums(/var RATIOS = \[([^\]]+)\]/, 'RATIOS')
if (ratios.length && !eq(ratios, HEADING_RATIOS)) {
  fails.push(`字阶比例不同步：\n  index.html    = [${ratios}]\n  typography.ts = [${HEADING_RATIOS}]`)
}

// 2. 可选字号档位
const sizes = nums(/var SIZES = \[([^\]]+)\]/, 'SIZES')
if (sizes.length && !eq(sizes, PROSE_SIZES)) {
  fails.push(`字号档位不同步：\n  index.html    = [${sizes}]\n  typography.ts = [${PROSE_SIZES}]`)
}

// 3. 默认字号
const def = /if \(SIZES\.indexOf\(size\) < 0\) size = (\d+)/.exec(html)
if (!def) fails.push('index.html 里找不到默认字号回落')
else if (Number(def[1]) !== PROSE_SIZE_DEFAULT) {
  fails.push(`默认字号不同步：index.html = ${def[1]}，typography.ts = ${PROSE_SIZE_DEFAULT}`)
}

// 4. 文章标题比例
const title = /'--type-title', Math\.round\(size \* ([\d.]+)\)/.exec(html)
if (!title) fails.push('index.html 里找不到 --type-title 计算')
else if (Number(title[1]) !== TITLE_RATIO) {
  fails.push(`标题比例不同步：index.html = ${title[1]}，typography.ts = ${TITLE_RATIO}`)
}

// 5. 字体栈：防闪脚本里的每个 id 都要能在 PROSE_FONTS 找到同样的栈
const fontBlock = /var FONTS = \{([\s\S]*?)\};/.exec(html)
if (!fontBlock) fails.push('index.html 里找不到 FONTS 表')
else {
  const inHtml = new Map<string, string>()
  for (const m of fontBlock[1].matchAll(/(\w+):\s*'([^']+)'/g)) inHtml.set(m[1], m[2])
  for (const f of PROSE_FONTS) {
    if (!f.stack) continue // system 项无需内联
    const got = inHtml.get(f.id)
    if (got === undefined) fails.push(`字体 "${f.id}" 在 index.html 的 FONTS 表里缺失`)
    else if (got !== f.stack) {
      fails.push(`字体 "${f.id}" 栈不同步：\n  index.html    = ${got}\n  typography.ts = ${f.stack}`)
    }
  }
  for (const id of inHtml.keys()) {
    if (!PROSE_FONTS.some((f) => f.id === id)) {
      fails.push(`字体 "${id}" 只存在于 index.html，typography.ts 里已无对应项`)
    }
  }
}

if (fails.length) {
  console.error('✗ 排版比例不同步：\n\n' + fails.join('\n\n'))
  process.exit(1)
}
console.log('✓ 排版比例同步：字阶 / 字号档位 / 默认值 / 标题比例 / 字体栈 均一致')
