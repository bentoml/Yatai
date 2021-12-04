import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllModelVersions } from '@/services/model_version'
import { usePage } from '@/hooks/usePage'
import { IModelVersionSchema } from '@/schemas/model_version'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import ModelVersionImageBuildStatusTag from '@/components/ModelVersionImageBuildStatus'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useQ } from '@/hooks/useQ'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import { ResourceLabels } from './ResourceLabels'

export default function ModelVersionFlatListCard() {
    const { q, updateQ } = useQ()
    const [page] = usePage()
    const queryKey = `fetchAllModelVersions:${qs.stringify(page)}`
    const modelVersionsInfo = useQuery(queryKey, () => listAllModelVersions(page))
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()

    const uids = useMemo(
        () => modelVersionsInfo.data?.items.map((modelVersion) => modelVersion.uid) ?? [],
        [modelVersionsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (modelVersion: IModelVersionSchema) => {
            queryClient.setQueryData(
                queryKey,
                (oldData?: IListSchema<IModelVersionSchema>): IListSchema<IModelVersionSchema> => {
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
                }
            )
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'model_version',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'model_version',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.model_version}
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
                    modelVersionsInfo.data?.items.map((modelVersion) => [
                        <div
                            key={modelVersion.uid}
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 10,
                            }}
                        >
                            <Link to={`/models/${modelVersion.model.name}/versions/${modelVersion.version}`}>
                                {modelVersion.model.name}:{modelVersion.version}
                            </Link>
                            <ResourceLabels resource={modelVersion} />
                        </div>,
                        <ModelVersionImageBuildStatusTag
                            key={modelVersion.uid}
                            status={modelVersion.image_build_status}
                        />,
                        modelVersion.description,
                        modelVersion.creator && <User user={modelVersion.creator} />,
                        formatDateTime(modelVersion.build_at),
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
