#!/usr/bin/env bash
# =============================================================================
# Gridea Pro — macOS DMG installer background generator
# -----------------------------------------------------------------------------
# 用 ImageMagick 生成一张 600x400 的安装背景图（@2x 渲染后下采样，更清晰）。
#
# 用法：
#   ./create-background.sh [output.png]
#
# 依赖：
#   ImageMagick v7（macOS GitHub runner 已预装；本地可 `brew install imagemagick`）
#
# 设计说明：
#   - 600x400 窗口尺寸，与 release.yml 中 create-dmg 的 --window-size 一致
#   - 图标位置：App 在 (150, 200)，Applications 在 (450, 200)
#   - 标题居顶，副标题在标题下方
#   - 中间一根紫色箭头从 App 指向 Applications，提示拖拽方向
# =============================================================================
set -euo pipefail

OUT="${1:-background.png}"

# 在 2x 画布上绘制后缩小，得到更锐利的文字与曲线
W=1200
H=800

magick -size ${W}x${H} \
  gradient:'#EEF2FF-#FFFFFF' \
  -font Helvetica-Bold -pointsize 64 -fill '#1E293B' \
    -gravity North -annotate +0+70 'Gridea Pro' \
  -font Helvetica -pointsize 26 -fill '#64748B' \
    -gravity North -annotate +0+170 'Drag the app into your Applications folder' \
  -stroke '#6366F1' -strokewidth 8 -strokelinecap round -fill none \
    -draw "line 460,400 740,400" \
  -fill '#6366F1' -stroke none \
    -draw "polygon 740,378 790,400 740,422" \
  -font Helvetica -pointsize 22 -fill '#94A3B8' \
    -gravity South -annotate +0+60 'gridea.pro' \
  -resize 600x400 \
  "$OUT"

echo "✔ Generated DMG background → $OUT"
