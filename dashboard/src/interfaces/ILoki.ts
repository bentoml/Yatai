export type LokiFilterType = 'contains' | 'not contains'

export interface ILokiLabelFilterNode {
    name: string
    value: string
    operator: '=' | '!=' | '=~' | '!~' | '<' | '<=' | '>' | '>='
}

export interface ILokiLineFilterNode {
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
