import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createYataiComponent } from '@/services/yatai_component'
import {
    ICreateYataiComponentSchema,
    IYataiComponentSchema,
    YataiComponentReleaseStatus,
} from '@/schemas/yatai_component'
import YataiComponentForm from '@/components/YataiComponentForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { IListSchema } from '@/schemas/list'
import { useSubscription } from '@/hooks/useSubscription'
import { StyledSpinnerNext } from 'baseui/spinner'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'

export interface IYataiComponentListCardProps {
    orgName: string
    clusterName: string
}

export default function YataiComponentListCard({ orgName, clusterName }: IYataiComponentListCardProps) {
    const { yataiComponentsInfo, queryKey } = useFetchYataiComponents(orgName, clusterName)
    const [isCreateYataiComponentOpen, setIsCreateYataiComponentOpen] = useState(false)
    const handleCreateYataiComponent = useCallback(
        async (data: ICreateYataiComponentSchema) => {
            await createYataiComponent(orgName, clusterName, data)
            await yataiComponentsInfo.refetch()
            setIsCreateYataiComponentOpen(false)
        },
        [orgName, clusterName, yataiComponentsInfo]
    )
    const [t] = useTranslation()

    const statusColorMap: Record<YataiComponentReleaseStatus, keyof TagKind> = useMemo(() => {
        return {
            'unknown': TagKind.black,
            'deployed': TagKind.positive,
            'running': TagKind.positive,
            'unhealthy': TagKind.warning,
            'failed': TagKind.negative,
            'deploying': TagKind.accent,
            'uninstalled': TagKind.orange,
            'superseded': TagKind.yellow,
            'uninstalling': TagKind.purple,
            'pending-install': TagKind.primary,
            'pending-upgrade': TagKind.primary,
            'pending-rollback': TagKind.primary,
        }
    }, [])

    const componentTypes = useMemo(
        () => yataiComponentsInfo.data?.map((yataiComponent) => yataiComponent.type) ?? [],
        [yataiComponentsInfo.data]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (yataiComponent: IYataiComponentSchema) => {
            queryClient.setQueryData(
                queryKey,
                (oldData?: IListSchema<IYataiComponentSchema>): IListSchema<IYataiComponentSchema> => {
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
                        items: oldData.items.map((oldYataiComponent) => {
                            if (oldYataiComponent.type === yataiComponent.type) {
                                return {
                                    ...oldYataiComponent,
                                    ...yataiComponent,
                                }
                            }
                            return oldYataiComponent
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
            resourceType: 'yatai_component',
            resourceUids: componentTypes,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'yatai_component',
                resourceUids: componentTypes,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, componentTypes, unsubscribe])

    return (
        <Card
            title={t('sth list', [t('yatai component')])}
            titleIcon={resourceIconMapping.yatai_component}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateYataiComponentOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={yataiComponentsInfo.isLoading}
                columns={[t('type'), t('status'), t('created_at')]}
                data={
                    yataiComponentsInfo.data?.map((yataiComponent) => [
                        <Link
                            key={yataiComponent.type}
                            to={`/orgs/${orgName}/clusters/${clusterName}/yatai_components/${yataiComponent.type}`}
                        >
                            {yataiComponent.type}
                        </Link>,
                        <Tag
                            key={yataiComponent.type}
                            closeable={false}
                            variant={TagVariant.light}
                            kind={
                                yataiComponent.release ? statusColorMap[yataiComponent.release.info.status] : undefined
                            }
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 4,
                                }}
                            >
                                {['deploying'].indexOf(yataiComponent.release?.info.status ?? '') >= 0 && (
                                    <StyledSpinnerNext $size={100} />
                                )}
                                {yataiComponent.release ? t(yataiComponent.release.info.status) : '-'}
                            </div>
                        </Tag>,
                        yataiComponent.release ? formatTime(yataiComponent.release.info.last_deployed) : '-',
                    ]) ?? []
                }
            />
            <Modal
                isOpen={isCreateYataiComponentOpen}
                onClose={() => setIsCreateYataiComponentOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('yatai component')])}</ModalHeader>
                <ModalBody>
                    <YataiComponentForm
                        onSubmit={handleCreateYataiComponent}
                        orgName={orgName}
                        clusterName={clusterName}
                    />
                </ModalBody>
            </Modal>
        </Card>
    )
}
