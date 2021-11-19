import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllBentoVersions } from '@/services/bento_version'
import { usePage } from '@/hooks/usePage'
import { IBentoVersionSchema } from '@/schemas/bento_version'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import BentoVersionImageBuildStatusTag from '@/components/BentoVersionImageBuildStatus'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'

export default function BentoVersionFlatListCard() {
    const { q, updateQ } = useQ()
    const [page] = usePage()
    const queryKey = `fetchAllBentoVersions:${qs.stringify(page)}`
    const bentoVersionsInfo = useQuery(queryKey, () => listAllBentoVersions(page))
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()

    const uids = useMemo(
        () => bentoVersionsInfo.data?.items.map((bentoVersion) => bentoVersion.uid) ?? [],
        [bentoVersionsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (bentoVersion: IBentoVersionSchema) => {
            queryClient.setQueryData(
                queryKey,
                (oldData?: IListSchema<IBentoVersionSchema>): IListSchema<IBentoVersionSchema> => {
                    if (!oldData) {
                        return {
                            start: 0,
                            count: 0,
                            total: 0,
                            items: [],
                        }
                    }
                    return {
                        ...oldData,
                        items: oldData.items.map((oldBentoVersion) => {
                            if (oldBentoVersion.uid === bentoVersion.uid) {
                                return {
                                    ...oldBentoVersion,
                                    ...bentoVersion,
                                }
                            }
                            return oldBentoVersion
                        }),
                    }
                }
            )
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'bento_version',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'bento_version',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.bento_version}
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
                                    label: t('the bentos I created'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
        >
            <FilterBar
                filters={[
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
                                id: 'build_at-desc',
                                label: t('newest build'),
                            },
                            {
                                id: 'build_at-asc',
                                label: t('oldest build'),
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
            <Table
                isLoading={bentoVersionsInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('build_at')]}
                data={
                    bentoVersionsInfo.data?.items.map((bentoVersion) => [
                        <Link
                            key={bentoVersion.uid}
                            to={`/bentos/${bentoVersion.bento.name}/versions/${bentoVersion.version}`}
                        >
                            {bentoVersion.bento.name}:{bentoVersion.version}
                        </Link>,
                        <BentoVersionImageBuildStatusTag
                            key={bentoVersion.uid}
                            status={bentoVersion.image_build_status}
                        />,
                        bentoVersion.description,
                        bentoVersion.creator && <User user={bentoVersion.creator} />,
                        formatDateTime(bentoVersion.build_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentoVersionsInfo.data?.start,
                    count: bentoVersionsInfo.data?.count,
                    total: bentoVersionsInfo.data?.total,
                    afterPageChange: () => {
                        bentoVersionsInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
