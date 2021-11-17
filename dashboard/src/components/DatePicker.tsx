import React from 'react'
import { DatePicker as BaseDatePicker } from 'baseui/datepicker'
import moment from 'moment'

interface IDatePickerProps {
    value?: string
    onChange?: (value?: string) => void
}

export default function DatePicker({ value, onChange }: IDatePickerProps) {
    return (
        <BaseDatePicker
            clearable
            value={value ? moment(value).toDate() : undefined}
            onChange={(e) => {
                const date = Array.isArray(e.date) ? e.date[0] : e.date
                onChange?.(date ? moment(date).startOf('day').toDate().toISOString() : undefined)
            }}
        />
    )
}
