export type LokiFilterType = 'contains' | 'not contains'

export interface ILokiFilter {
    type: LokiFilterType
    isRegexp: boolean
    value: string
}

export interface ILokiLogDownload {
    query?: string
    start?: string
    end?: string
    follow?: boolean
}
