/* eslint-disable prettier/prettier */
import React, { useEffect } from 'react'
import { usePage } from '@/hooks/usePage'
import useTranslation from '@/hooks/useTranslation'
import { listOrganizationEventOperationNames, listOrganizationEvents } from '@/services/organization'
import qs from 'qs'
import { useQuery } from 'react-query'
import { FiActivity } from 'react-icons/fi'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import { DatePicker } from 'baseui/datepicker'
import moment from 'moment'
import { ResourceType } from '@/schemas/resource'
import { resourceIconMapping } from '@/consts'
import Card from './Card'
import EventList from './EventList'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import User from './User'

export default function EventListCard() {
    const [page] = usePage()
    const { q, updateQ } = useQ()
    const membersInfo = useFetchOrganizationMembers()
    const eventsInfo = useQuery(`listOrganizationEvents:${qs.stringify(page)}`, () => listOrganizationEvents(page))
    const [t] = useTranslation()
    const [rangeDate, setRangeDate] = React.useState<Date[] | null>(
        q.started_at && q.ended_at
            ? [moment((q.started_at as string[])[0]).toDate(), moment((q.ended_at as string[])[0]).toDate()]
            : null
    )
    const operationNamesInfo = useQuery(`listOrganizationOperationNames:${q.resource_type ? (q.resource_type as string[])[0] : ''}`, () => q.resource_type ? listOrganizationEventOperationNames((q.resource_type as string[])[0] as ResourceType) : Promise.resolve([]))

    useEffect(() => {
        if (!rangeDate) {
            return
        }
        if (!rangeDate[0] || !rangeDate[1]) {
            updateQ({
                started_at: undefined,
                ended_at: undefined,
            })
            return
        }
        updateQ({
            started_at: rangeDate && [moment(rangeDate[0]).format('YYYY-MM-DD')],
            ended_at: rangeDate && [moment(rangeDate[1]).format('YYYY-MM-DD')],
        })
    }, [rangeDate, updateQ])

    return (
        <Card
            title={t('events')}
            titleIcon={FiActivity}
            middle={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexGrow: 1,
                    }}
                >
                    <div
                        style={{
                            width: 100,
                            flexGrow: 1,
                        }}
                    />
                    <div
                        style={{
                            flexGrow: 2,
                            flexShrink: 0,
                            maxWidth: 1200,
                        }}
                    >
                        <FilterInput
                            filterConditions={[
                                {
                                    qStr: 'creator:@me',
                                    label: t('the events I created'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
        >
            <FilterBar
                prefix={
                    <DatePicker
                        range
                        clearable
                        value={rangeDate}
                        onChange={({ date }) => setRangeDate(date as Date[])}
                        placeholder='YYYY/MM/DD – YYYY/MM/DD'
                        quickSelect
                        size='compact'
                    />
                }
                filters={[
                    {
                        showInput: true,
                        multiple: false,
                        options: [
                            {
                                id: 'bento',
                                label: <div style={{ display: 'flex', alignItems: 'center', gap: 6, }}>{React.createElement(resourceIconMapping.bento, { size: 12 })}{t('bento')}</div>,
                            },
                            {
                                id: 'model',
                                label: <div style={{ display: 'flex', alignItems: 'center', gap: 6, }}>{React.createElement(resourceIconMapping.model, { size: 12 })}{t('model')}</div>,
                            },
                            {
                                id: 'deployment',
                                label: <div style={{ display: 'flex', alignItems: 'center', gap: 6, }}>{React.createElement(resourceIconMapping.deployment, { size: 12 })}{t('deployment')}</div>,
                            },
                        ],
                        value: ((q.resource_type as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                resource_type: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('resource type'),
                    },
                    {
                        showInput: true,
                        multiple: false,
                        options:
                            operationNamesInfo.data?.map((operationName) => ({
                                id: operationName,
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                label: t(operationName as any),
                            })) ?? [],
                        value: ((q.operation_name as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                operation_name: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('operation name'),
                    },
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],
                        value: ((q.creator as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                creator: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('creator'),
                    },
                    {
                        options: [
                            {
                                id: 'created_at-desc',
                                label: t('newest create'),
                            },
                            {
                                id: 'created_at-asc',
                                label: t('oldest create'),
                            },
                        ],
                        value: ((q.sort as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                sort: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('sort'),
                    },
                ]}
            />
            <EventList
                events={eventsInfo.data?.items ?? []}
                isLoading={eventsInfo.isLoading}
                paginationProps={{
                    start: eventsInfo.data?.start,
                    count: eventsInfo.data?.count,
                    total: eventsInfo.data?.total,
                    afterPageChange: () => {
                        eventsInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}

