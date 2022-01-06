import { StatefulTooltip } from 'baseui/tooltip'
import React from 'react'
import ReactTimeAgo from 'react-time-ago'
import { formatDateTime } from '@/utils/datetime'
import i18n from '@/i18n'

export interface ITimeProps {
    time: string
}

export default function Time({ time }: ITimeProps) {
    return (
        <StatefulTooltip placement='bottom' content={() => formatDateTime(time)}>
            <ReactTimeAgo date={new Date(time)} timeStyle='round' locales={i18n.languages as string[]} />
        </StatefulTooltip>
    )
}
