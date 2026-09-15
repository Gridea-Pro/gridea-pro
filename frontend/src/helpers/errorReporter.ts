import { safeEventsEmit } from '@/helpers/wailsRuntime'

/** 错误来源：用于判断该以哪种方式呈现，以及日志里区分归类 */
export type ErrorSource = 'vue' | 'window' | 'promise' | 'boundary'

export interface ReportedError {
  source: ErrorSource
  message: string
  stack: string
  /** 出错位置的补充说明：Vue 组件链、路由路径等 */
  context: string
  time: string
}

function toMessage(err: unknown): string {
  if (err instanceof Error) return err.message || err.name
  if (typeof err === 'string') return err
  try {
    return JSON.stringify(err)
  } catch {
    return String(err)
  }
}

function toStack(err: unknown): string {
  return err instanceof Error && err.stack ? err.stack : ''
}

/**
 * 统一收口前端错误：写控制台 + 发给 Go 侧落盘。
 * 呈现由调用方决定（ErrorBoundary 局部降级 / toast），本函数不碰 UI。
 */
export function reportError(err: unknown, source: ErrorSource, context = ''): ReportedError {
  const reported: ReportedError = {
    source,
    message: toMessage(err),
    stack: toStack(err),
    context,
    time: new Date().toISOString(),
  }

  console.error(`❌ [${source}]${context ? ` ${context}` : ''}`, err)

  // Go 侧监听 renderer-error 写入日志文件，便于用户反馈问题时附带
  safeEventsEmit('renderer-error', JSON.stringify(reported))

  return reported
}

/** 供「复制诊断信息」按钮使用：拼成可直接贴进 issue 的纯文本 */
export function formatDiagnostics(e: ReportedError): string {
  return [
    `时间: ${e.time}`,
    `来源: ${e.source}`,
    e.context ? `位置: ${e.context}` : '',
    `路由: ${window.location.hash || window.location.pathname}`,
    `UA: ${navigator.userAgent}`,
    '',
    `错误: ${e.message}`,
    '',
    e.stack || '(无堆栈)',
  ]
    .filter(Boolean)
    .join('\n')
}
