/** 部署平台类型 */
export type PlatformType = 'github' | 'coding' | 'sftp' | 'netlify' | 'vercel'

/** 系统设置 — 与后端 domain.Setting 一一对应 */
export interface ISetting {
  platform: PlatformType
  platformConfigs: Record<string, Record<string, any>>
  enabledPlatforms?: string[]
  proxyEnabled?: boolean
  proxyURL?: string
}

/** 设置表单（BasicSetting 内部使用的 UI 层平铺结构） */
export interface ISettingForm {
  platform: PlatformType
  domain: string
  repository: string
  branch: string
  username: string
  email: string
  tokenUsername: string
  token: string
  cname: string
  transferProtocol: string
  port: string
  server: string
  password: string
  privateKey: string
  remotePath: string
  netlifyAccessToken: string
  netlifySiteId: string
  proxyEnabled: boolean
  proxyURL: string
  [index: string]: any
}

export interface ICommentSetting {
  showComment: boolean
  commentPlatform: string
  gitalkSetting?: any
  disqusSetting?: any
  [key: string]: any
}
