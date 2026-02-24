/**
 * 文章统计 Composable
 * 
 * 计算文章字数、阅读时间等统计信息
 */

import { computed } from 'vue'
import { wordCount, timeCalc } from '@/helpers/words-count'
import type { ArticleStats } from '@/types/post'

export function useArticleStats(content: () => string) {
    /**
     * 计算文章统计信息
     */
    const stats = computed<ArticleStats>(() => {
        const contentValue = content()

        // 计算阅读时间
        const reading = timeCalc(contentValue)
        const seconds = Number((reading.second - (reading.minius - 1) * 60).toFixed(2))
        const minutes = Math.floor(reading.second / 60)

        // 格式化时间字符串
        const formatTime = seconds < 60
            ? `${minutes}m ${seconds}s`
            : `${minutes}m`

        // 计算字数
        let wordsNumber = 0
        wordCount(contentValue, (count: number) => {
            wordsNumber = count
        })

        return {
            formatTime,
            wordsNumber: Array.isArray(wordsNumber) ? 0 : wordsNumber
        }
    })

    /**
     * 字数
     */
    const wordsNumber = computed(() => stats.value.wordsNumber)

    /**
     * 阅读时间（格式化字符串）
     */
    const readingTime = computed(() => stats.value.formatTime)

    /**
     * 预估阅读分钟数
     */
    const readingMinutes = computed(() => {
        const contentValue = content()
        const reading = timeCalc(contentValue)
        return Math.ceil(reading.second / 60)
    })

    return {
        stats,
        wordsNumber,
        readingTime,
        readingMinutes
    }
}
