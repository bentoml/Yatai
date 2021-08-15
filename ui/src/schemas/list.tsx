export interface IListSchema<T> {
    total: number
    start: number
    count: number
    items: T[]
}

export interface IListQuerySchema {
    start: number
    count: number
    search?: string
    sort_by?: string
    sort_asc?: boolean
}
