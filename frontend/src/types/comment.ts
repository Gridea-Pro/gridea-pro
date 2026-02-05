export interface Comment {
    id: string
    avatar: string
    nickname: string
    url?: string
    content: string
    createdAt: string
    articleTitle: string
    articleId: string
    articleUrl?: string
    isNew?: boolean
    parentNick?: string
}

export interface CommentSettings {
    enable: boolean
    platform: 'Valine' | 'Waline' | 'Twikoo' | 'Gitalk' | 'Disqus' | 'Cusdis' | 'Giscus'
    platformConfigs: Record<string, Record<string, any>>
}
