import moment from 'moment'
import { dateTimeFormat } from '@/consts'

export function formatDateTime(s: string, format = 'YYYY-MM-DDTHH:mm:ssZ'): string {
    return moment(s, format).format(dateTimeFormat)
}

export function durationToStr(v: number) {
    const units = ['μs', 'ms', 's', 'm', 'h', 'd']
    let basic = 1000
    let unitIdx = 0
    let newV = v
    while (newV >= basic) {
        unitIdx++
        newV /= basic
        if (unitIdx > 2) {
            basic = 60
        }
        if (unitIdx > 4) {
            basic = 24
        }
    }
    return `${newV.toFixed(2)}${units[unitIdx]}`
}
