import { useSubscription } from '@/hooks/useSubscription'
import useTranslation from '@/hooks/useTranslation'
import { IPaginationProps } from '@/interfaces/IPaginationProps'
import { IBentoSchema, IBentoWithRepositorySchema } from '@/schemas/bento'
import { IListSchema } from '@/schemas/list'
import { recreateBentoImageBuilderJob } from '@/services/bento'
import { ListItem } from 'baseui/list'
import prettyBytes from 'pretty-bytes'
import React, { useCallback, useEffect, useMemo } from 'react'
import { useQueryClient } from 'react-query'
import { Link } from 'react-router-dom'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import List from './List'
import { ResourceLabels } from './ResourceLabels'
import Time from './Time'
import User from './User'

export interface IBentoListProps {
    queryKey: string
    isLoading: boolean
    bentos: IBentoWithRepositorySchema[]
    paginationProps?: IPaginationProps
}

export default function BentoList({ queryKey, isLoading, bentos, paginationProps }: IBentoListProps) {
    const [t] = useTranslation()

    const uids = useMemo(() => bentos.map((bento) => bento.uid) ?? [], [bentos])
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (bentoVersion: IBentoSchema) => {
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
                            podsSelector={`yatai.io/bento=${bento.version},yatai.io/bento-repository=${bento.repository.name}`}
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
                                    justifyContent: 'flex-end',
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
                        <Link to={`/bento_repositories/${bento.repository.name}/bentos/${bento.version}`}>
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
        <List isLoading={isLoading} items={bentos} onRenderItem={handleRenderItem} paginationProps={paginationProps} />
    )
}
