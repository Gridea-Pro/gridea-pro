export interface IThemeConfigItemOption {
  value: string | number | boolean
  label?: string
}

export interface IThemeConfigArrayField {
  name: string
  type: string
  label?: string
  note?: string
  card?: string
  options?: IThemeConfigItemOption[]
}

export interface IThemeConfigItem {
  name: string
  type: string
  label?: string
  group?: string
  note?: string
  card?: string
  value?: unknown
  options?: IThemeConfigItemOption[]
  arrayItems?: IThemeConfigArrayField[]
}

export interface ITheme {
  themeName: string
  postPageSize: number
  archivesPageSize: number
  siteName: string
  siteAuthor: string
  siteEmail?: string
  siteDescription: string
  footerInfo: string
  postUrlFormat: string
  tagUrlFormat: string
  dateFormat: string
  language: string
  feedEnabled: boolean
  feedFullText: boolean
  feedCount: number
  postPath: string
  tagPath: string
}
