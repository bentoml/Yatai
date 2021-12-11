import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllBentos, recreateBentoImageBuilderJob } from '@/services/bento'
import { usePage } from '@/hooks/usePage'
import { IBentoSchema } from '@/schemas/bento'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import ImageBuildStatusTag from '@/components/ImageBuildStatusTag'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import { ResourceLabels } from './ResourceLabels'

export default function BentoFlatListCard() {
    const { q, updateQ } = useQ()
    const [page] = usePage()
    const queryKey = `fetchAllBentos:${qs.stringify(page)}`
    const bentosInfo = useQuery(queryKey, () => listAllBentos(page))
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()

    const uids = useMemo(
        () => bentosInfo.data?.items.map((bentoVersion) => bentoVersion.uid) ?? [],
        [bentosInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (bento: IBentoSchema) => {
            queryClient.setQueryData(queryKey, (oldData?: IListSchema<IBentoSchema>): IListSchema<IBentoSchema> => {
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
                    items: oldData.items.map((oldBento) => {
                        if (oldBento.uid === bento.uid) {
                            return {
                                ...oldBento,
                                ...bento,
                            }
                        }
                        return oldBento
                    }),
                }
            })
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'bento',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'bento',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('bento')])}
            titleIcon={resourceIconMapping.bento}
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
                isLoading={bentosInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('build_at')]}
                data={
                    bentosInfo.data?.items.map((bento) => [
                        <div
                            key={bento.uid}
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 10,
                            }}
                        >
                            <Link to={`/bento_repositories/${bento.repository.name}/bentos/${bento.version}`}>
                                {bento.repository.name}:{bento.version}
                            </Link>
                            <ResourceLabels resource={bento} />
                        </div>,
                        <ImageBuildStatusTag
                            key={bento.uid}
                            status={bento.image_build_status}
                            podsSelector={`yatai.io/bento=${bento.version},yatai.io/bento-repository=${
                                bento.repository.name
                            };yatai.io/model in (${bento.manifest.models.map((x) => x.split(':')[1]).join(',')})`}
                            onRerunClick={async () => {
                                await recreateBentoImageBuilderJob(bento.repository.name, bento.version)
                            }}
                        />,
                        bento.description,
                        bento.creator && <User user={bento.creator} />,
                        formatDateTime(bento.build_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentosInfo.data?.start,
                    count: bentosInfo.data?.count,
                    total: bentosInfo.data?.total,
                    afterPageChange: () => {
                        bentosInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
