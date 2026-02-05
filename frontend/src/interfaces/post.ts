export interface IPostData {
  title: string
  date: string
  published: boolean
  hideInList: boolean
  tags?: string[]
  categories?: string[]
  feature: string
  isTop: boolean
}

export interface IPost {
  content: string

  data: IPostData

  fileName: string
}
