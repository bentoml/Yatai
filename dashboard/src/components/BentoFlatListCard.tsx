import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllBentos, recreateBentoImageBuilderJob } from '@/services/bento'
import { usePage } from '@/hooks/usePage'
import { IBentoSchema, IBentoWithRepositorySchema } from '@/schemas/bento'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import prettyBytes from 'pretty-bytes'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import { ResourceLabels } from './ResourceLabels'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import Time from './Time'
import List from './List'
import Link from './Link'
import ListItem from './ListItem'

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

    const handleRenderItem = useCallback(
        (bento: IBentoWithRepositorySchema) => {
            return (
                <ListItem
                    key={bento.uid}
                    artwork={() => (
                        <ImageBuildStatusIcon
                            key={bento.uid}
                            status={bento.image_build_status}
                            podsSelector={`yatai.ai/bento=${bento.version},yatai.ai/bento-repository=${bento.repository.name}`}
                            onRerunClick={async () => {
                                await recreateBentoImageBuilderJob(bento.repository.name, bento.version)
                            }}
                        />
                    )}
                    endEnhancer={() => (
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 8,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 8,
                                }}
                            >
                                <div>{prettyBytes(bento.manifest.size_bytes)}</div>
                            </div>
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 8,
                                }}
                            >
                                {bento.creator && <User size='16px' user={bento.creator} />}
                                {t('Created At')}
                                <Time time={bento.created_at} />
                            </div>
                        </div>
                    )}
                >
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 10,
                        }}
                    >
                        <Link href={`/bento_repositories/${bento.repository.name}/bentos/${bento.version}`}>
                            {bento.repository.name}:{bento.version}
                        </Link>
                        <ResourceLabels resource={bento} />
                    </div>
                </ListItem>
            )
        },
        [t]
    )

    return (
        <Card
            title={t('bentos')}
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
            <List
                isLoading={bentosInfo.isLoading}
                items={bentosInfo.data?.items ?? []}
                onRenderItem={handleRenderItem}
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
