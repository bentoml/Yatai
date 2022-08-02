import moment from 'moment'
import { dateTimeFormat } from '@/consts'

export function timeStrToMoment(timeStr: string, format = 'YYYY-MM-DDTHH:mm:ssZ'): moment.Moment | null {
    if (timeStr) {
        return moment(timeStr, format)
    }
    return null
}

export function formatMoment(m: moment.Moment): string {
    return m.format(dateTimeFormat)
}

export function formatDateTime(s: string, format = 'YYYY-MM-DDTHH:mm:ssZ'): string {
    const m = timeStrToMoment(s, format)
    if (m) {
        return m.format(dateTimeFormat)
    }
    return s
}
