import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllModels } from '@/services/model'
import { usePage } from '@/hooks/usePage'
import { IModelSchema } from '@/schemas/model'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import { ResourceLabels } from './ResourceLabels'
import ImageBuildStatusTag from './ImageBuildStatusTag'

export default function ModelFlatListCard() {
    const { q, updateQ } = useQ()
    const [page] = usePage()
    const queryKey = `fetchAllModels:${qs.stringify(page)}`
    const modelVersionsInfo = useQuery(queryKey, () => listAllModels(page))
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()

    const uids = useMemo(
        () => modelVersionsInfo.data?.items.map((modelVersion) => modelVersion.uid) ?? [],
        [modelVersionsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (modelVersion: IModelSchema) => {
            queryClient.setQueryData(queryKey, (oldData?: IListSchema<IModelSchema>): IListSchema<IModelSchema> => {
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
                    items: oldData.items.map((oldModelVersion) => {
                        if (oldModelVersion.uid === modelVersion.uid) {
                            return {
                                ...oldModelVersion,
                                ...modelVersion,
                            }
                        }
                        return oldModelVersion
                    }),
                }
            })
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'model',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'model',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('model')])}
            titleIcon={resourceIconMapping.model}
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
                                    label: t('the models I created'),
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
                isLoading={modelVersionsInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('build_at')]}
                data={
                    modelVersionsInfo.data?.items.map((model) => [
                        <div
                            key={model.uid}
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 10,
                            }}
                        >
                            <Link to={`/model_repositories/${model.repository.name}/models/${model.version}`}>
                                {model.repository.name}:{model.version}
                            </Link>
                            <ResourceLabels resource={model} />
                        </div>,
                        <ImageBuildStatusTag
                            key={model.uid}
                            status={model.image_build_status}
                            podsSelector={`yatai.io/model=${model.version},yatai.io/model-repository=${model.repository.name}`}
                        />,
                        model.description,
                        model.creator && <User user={model.creator} />,
                        formatDateTime(model.build_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: modelVersionsInfo.data?.start,
                    count: modelVersionsInfo.data?.count,
                    total: modelVersionsInfo.data?.total,
                    afterPageChange: () => {
                        modelVersionsInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
