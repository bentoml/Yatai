import { useCallback, useEffect, useMemo } from 'react'
import { useQueryClient } from 'react-query'
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
import Time from './Time'
import Link from './Link'
import ListItem from './ListItem'

export interface IModelListProps {
    queryKey: string
    isLoading: boolean
    models: IModelWithRepositorySchema[]
    paginationProps?: IPaginationProps
    isListItem?: boolean
    size?: 'default' | 'small'
}

export default function ModelList({
    queryKey,
    isLoading,
    models,
    paginationProps,
    isListItem = true,
    size = 'default',
}: IModelListProps) {
    const [t] = useTranslation()

    const uids = useMemo(() => models?.map((modelVersion) => modelVersion.uid) ?? [], [models])
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (modelVersion: IModelSchema) => {
            if (isListItem) {
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
                        items:
                            oldData.items?.map((oldModelVersion) => {
                                if (oldModelVersion.uid === modelVersion.uid) {
                                    return {
                                        ...oldModelVersion,
                                        ...modelVersion,
                                    }
                                }
                                return oldModelVersion
                            }) ?? [],
                    }
                })
                return
            }
            queryClient.setQueryData(queryKey, (oldData?: IModelSchema[]): IModelSchema[] => {
                if (!oldData) {
                    return []
                }
                return (
                    oldData?.map((oldModelVersion) => {
                        if (oldModelVersion.uid === modelVersion.uid) {
                            return {
                                ...oldModelVersion,
                                ...modelVersion,
                            }
                        }
                        return oldModelVersion
                    }) ?? []
                )
            })
        },
        [isListItem, queryClient, queryKey]
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
                    overrides={
                        size === 'small'
                            ? {
                                  Content: {
                                      style: {
                                          padding: '0px',
                                          minHeight: '32px',
                                      },
                                  },
                              }
                            : undefined
                    }
                    artworkSize={size === 'small' ? 'SMALL' : undefined}
                    endEnhancer={
                        size === 'default'
                            ? () => {
                                  return (
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
                                  )
                              }
                            : undefined
                    }
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
                        {model.labels && model.labels.length > 0 && <ResourceLabels resource={model} />}
                    </div>
                </ListItem>
            )
        },
        [history, size, t]
    )

    return (
        <List isLoading={isLoading} items={models} onRenderItem={handleRenderItem} paginationProps={paginationProps} />
    )
}
