export interface ISetting {
  platform: 'github' | 'coding' | 'sftp' | 'gitee' | 'netlify'
  domain: string
  repository: string
  branch: string
  username: string
  email: string
  tokenUsername: string
  token: string
  cname: string
  port: string
  server: string
  password: string
  privateKey: string
  remotePath: string
  proxyPath: string
  proxyPort: string
  enabledProxy: 'direct' | 'proxy'
  netlifyAccessToken: string
  netlifySiteId: string
  [index: string]: string
}

export interface ICommentSetting {
  showComment: boolean
  commentPlatform: string
  gitalkSetting?: any
  disqusSetting?: any
  [key: string]: any
}


