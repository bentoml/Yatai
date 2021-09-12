export interface IKubeEventSchema {
    message: string
    reason: string
    involvedObject?: {
        kind: string
        name: string
    }
    type: 'Normal' | 'Warning'
    firstTimestamp: string
    lastTimestamp: string
}
