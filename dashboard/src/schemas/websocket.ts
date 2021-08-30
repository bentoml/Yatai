export interface IWsReqSchema<T> {
    type: 'data' | 'heartbeat'
    payload: T
}

export interface IWsRespSchema<T> {
    type: 'success' | 'error'
    message: string
    payload: T
}
