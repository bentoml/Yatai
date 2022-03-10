import { useCallback, useEffect, useMemo } from 'react'
import { useQueryClient } from 'react-query'
import { recreateModelImageBuilderJob } from '@/services/model'
import { IModelSchema, IModelWithRepositorySchema } from '@/schemas/model'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import prettyBytes from 'pretty-bytes'
import { MonoParagraphXSmall } from 'baseui/typography'
import { useHistory } from 'react-router-dom'
import { IPaginationProps } from '@/interfaces/IPaginationProps'
import { ResourceLabels } from './ResourceLabels'
import List from './List'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import Time from './Time'
import Link from './Link'
import ListItem from './ListItem'

export interface IModelListProps {
    queryKey: string
    isLoading: boolean
    models: IModelWithRepositorySchema[]
    paginationProps?: IPaginationProps
}

export default function ModelList({ queryKey, isLoading, models, paginationProps }: IModelListProps) {
    const [t] = useTranslation()

    const uids = useMemo(() => models.map((modelVersion) => modelVersion.uid) ?? [], [models])
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
                    items: oldData.items?.map((oldModelVersion) => {
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

    const history = useHistory()

    const handleRenderItem = useCallback(
        (model: IModelWithRepositorySchema) => {
            return (
                <ListItem
                    onClick={() => {
                        history.push(`/model_repositories/${model.repository.name}/models/${model.version}`)
                    }}
                    key={model.uid}
                    artwork={() => (
                        <ImageBuildStatusIcon
                            key={model.uid}
                            status={model.image_build_status}
                            podsSelector={`yatai.io/model=${model.version},yatai.io/model-repository=${model.repository.name}`}
                            onRerunClick={async () => {
                                await recreateModelImageBuilderJob(model.repository.name, model.version)
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
                                    gap: 20,
                                    justifyContent: 'flex-end',
                                }}
                            >
                                <div>{prettyBytes(model.manifest.size_bytes)}</div>
                                <div>{model.manifest.module}</div>
                            </div>
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 8,
                                }}
                            >
                                {model.creator && <User size='16px' user={model.creator} />}
                                {t('Created At')}
                                <Time time={model.created_at} />
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
                        <Link href={`/model_repositories/${model.repository.name}/models/${model.version}`}>
                            <MonoParagraphXSmall
                                overrides={{
                                    Block: {
                                        style: {
                                            margin: 0,
                                        },
                                    },
                                }}
                            >
                                {model.repository.name}:{model.version}
                            </MonoParagraphXSmall>
                        </Link>
                        <ResourceLabels resource={model} />
                    </div>
                </ListItem>
            )
        },
        [history, t]
    )

    return (
        <List isLoading={isLoading} items={models} onRenderItem={handleRenderItem} paginationProps={paginationProps} />
    )
}
