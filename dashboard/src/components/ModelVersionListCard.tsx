import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listModelVersions } from '@/services/model_version'
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

export interface IModelVersionListCardProps {
    modelName: string
}

export default function ModelVersionListCard({ modelName }: IModelVersionListCardProps) {
    const [page] = usePage()
    const queryKey = `fetchModelVersions:${modelName}:${qs.stringify(page)}`
    const modelVersionsInfo = useQuery(queryKey, () => listModelVersions(modelName, page))
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
        <Card title={t('sth list', [t('version')])} titleIcon={resourceIconMapping.model_version}>
            <Table
                isLoading={modelVersionsInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('created_at')]}
                data={
                    modelVersionsInfo.data?.items.map((modelVersion) => [
                        <Link key={modelVersion.uid} to={`/models/${modelName}/versions/${modelVersion.version}`}>
                            {modelVersion.version}
                        </Link>,
                        <ModelVersionImageBuildStatusTag
                            key={modelVersion.uid}
                            status={modelVersion.image_build_status}
                        />,
                        modelVersion.description,
                        modelVersion.creator && <User user={modelVersion.creator} />,
                        formatDateTime(modelVersion.created_at),
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
