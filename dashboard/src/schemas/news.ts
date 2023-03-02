export interface INewsItem {
    level?: 'info' | 'positive' | 'warning' | 'negative'
    title: string
    link: string
    cover?: string
    started_at?: string
    ended_at?: string
    version_constraint?: string
}

export interface INewsContent {
    notifications: INewsItem[]
    documentations: INewsItem[]
    release_notes: INewsItem[]
    blog_posts: INewsItem[]
}
