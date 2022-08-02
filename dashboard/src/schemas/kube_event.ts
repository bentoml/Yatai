import { timeStrToMoment } from '@/utils/datetime'
import moment from 'moment'

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
    eventTime: string
}

export function getEventTime(event: IKubeEventSchema): moment.Moment | null {
    if (event.eventTime) {
        return timeStrToMoment(event.eventTime)
    }
    if (event.lastTimestamp) {
        return timeStrToMoment(event.lastTimestamp)
    }
    if (event.firstTimestamp) {
        return timeStrToMoment(event.firstTimestamp)
    }
    return null
}
