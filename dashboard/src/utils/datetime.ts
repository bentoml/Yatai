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

export const dSecondsNormalized = (seconds: number) => {
    const normalizers = {
        seconds: 1,
        mins: 60,
        hours: 3600,
        days: 86400,
    }

    let str = ''
    Object.entries(normalizers).forEach(([key, value]) => {
        const div = seconds / value
        if (div > 1) {
            str = `${div.toFixed(0)} ${key}`
        }
    })

    return str
}
