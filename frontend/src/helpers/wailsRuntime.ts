/**
 * Wails Runtime 容错工具
 * 用于在纯 Vite 开发服务器环境下安全调用 Wails Runtime API
 */

import { EventsOn as WailsEventsOn, EventsEmit as WailsEventsEmit, WindowShow as WailsWindowShow } from '@/wailsjs/runtime'

/**
 * 检测当前是否运行在 Wails 环境中
 */
export function isWailsEnvironment(): boolean {
    if (typeof window === 'undefined') return false

    // 检测 Wails 注入的对象
    // @ts-ignore
    if (window.go || window.wails || window.runtime) return true

    // 检测 Wails Runtime 是否可用
    try {
        // @ts-ignore
        if (typeof window.runtime?.EventsOn === 'function') return true
    } catch {
        // ignore
    }

    return false
}

/**
 * 安全调用 EventsOn
 * 在非 Wails 环境下不会抛出错误
 */
export function safeEventsOn(eventName: string, callback: (...data: any[]) => void): (() => void) | null {
    if (!isWailsEnvironment()) {
        if (import.meta.env.DEV) {
            console.log(`[Wails Mock] EventsOn 注册: ${eventName} (已忽略)`)
        }
        return null
    }

    try {
        return WailsEventsOn(eventName, callback)
    } catch (error) {
        console.warn(`[Wails] EventsOn 失败: ${eventName}`, error)
        return null
    }
}

/**
 * 安全调用 EventsEmit
 * 在非 Wails 环境下不会抛出错误
 */
export function safeEventsEmit(eventName: string, ...data: any[]): void {
    if (!isWailsEnvironment()) {
        if (import.meta.env.DEV) {
            console.log(`[Wails Mock] EventsEmit: ${eventName}`, ...data)
        }
        return
    }

    try {
        WailsEventsEmit(eventName, ...data)
    } catch (error) {
        console.warn(`[Wails] EventsEmit 失败: ${eventName}`, error)
    }
}

/**
 * 安全调用 WindowShow
 * 在非 Wails 环境下不会抛出错误
 */
export function safeWindowShow(): void {
    if (!isWailsEnvironment()) {
        if (import.meta.env.DEV) {
            console.log('[Wails Mock] WindowShow (已忽略)')
        }
        return
    }

    try {
        WailsWindowShow()
    } catch (error) {
        console.warn('[Wails] WindowShow 失败:', error)
    }
}
