import { createAvatar } from '@dicebear/core'
import { micah } from '@dicebear/collection'

/**
 * 基于种子生成一致的头像
 * @param seed 种子值（用户名、邮箱、ID等）
 * @returns Data URL 格式的 SVG 头像，可直接用于 img 标签的 src 属性
 */
export function generateAvatar(seed: string): string {
    if (!seed) {
        seed = Math.random().toString(36).substring(7)
    }

    // 使用 micah 风格生成头像
    const avatar = createAvatar(micah, {
        seed,
        flip: false,  // 禁用翻转
        backgroundColor: ['B6E3F4', 'C0AEDC', 'D1D4F9', 'FFD5DC', 'FFDFBF'],
        backgroundType: ['gradientLinear'],  // 使用渐变线性背景
        backgroundRotation: [0, 30, 60, 90, 120, 150, 180, 210, 240, 270, 300, 330],
        baseColor: ['AC6651', 'F9C9B6'],
        facialHairColor: ['transparent'],
    })

    // 返回 Data URL 格式，可直接用于 <img src="">
    return avatar.toDataUri()
}
