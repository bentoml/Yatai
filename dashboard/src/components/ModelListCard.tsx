import { useCallback, useEffect, useMemo } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listModels } from '@/services/model'
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
import ImageBuildStatusTag from './ImageBuildStatusTag'

export interface IModelListCardProps {
    modelRepositoryName: string
}

export default function ModelListCard({ modelRepositoryName }: IModelListCardProps) {
    const [page] = usePage()
    const queryKey = `fetchModels:${modelRepositoryName}:${qs.stringify(page)}`
    const modelsInfo = useQuery(queryKey, () => listModels(modelRepositoryName, page))
    const [t] = useTranslation()

    const uids = useMemo(() => modelsInfo.data?.items.map((model) => model.uid) ?? [], [modelsInfo.data?.items])
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
        <Card title={t('models')} titleIcon={resourceIconMapping.model}>
            <Table
                isLoading={modelsInfo.isLoading}
                columns={[t('name'), t('image build status'), t('description'), t('creator'), t('created_at')]}
                data={
                    modelsInfo.data?.items.map((model) => [
                        <Link key={model.uid} to={`/model_repositories/${modelRepositoryName}/models/${model.version}`}>
                            {model.version}
                        </Link>,
                        <ImageBuildStatusTag key={model.uid} status={model.image_build_status} />,
                        model.description,
                        model.creator && <User user={model.creator} />,
                        formatDateTime(model.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: modelsInfo.data?.start,
                    count: modelsInfo.data?.count,
                    total: modelsInfo.data?.total,
                    afterPageChange: () => {
                        modelsInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
