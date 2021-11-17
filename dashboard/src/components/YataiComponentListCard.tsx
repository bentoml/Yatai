import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { createYataiComponent, deleteYataiComponent } from '@/services/yatai_component'
import {
    ICreateYataiComponentSchema,
    IYataiComponentSchema,
    YataiComponentReleaseStatus,
    YataiComponentType,
} from '@/schemas/yatai_component'
import YataiComponentForm from '@/components/YataiComponentForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Modal, ModalHeader, ModalBody, ModalFooter, ModalButton } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { Tag, KIND as TagKind, VARIANT as TagVariant } from 'baseui/tag'
import { IListSchema } from '@/schemas/list'
import { useSubscription } from '@/hooks/useSubscription'
import { StyledSpinnerNext } from 'baseui/spinner'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import { useFetchYataiComponentOperatorHelmCharts } from '@/hooks/useFetchYataiComponentOperatorHelmCharts'
import semver from 'semver'
import { useStyletron } from 'baseui'
import YataiComponentTypeRender from './YataiComponentTypeRender'
import { YataiComponentPodStatuses } from './YataiComponentPodStatuses'

export interface IYataiComponentListCardProps {
    clusterName: string
}

export default function YataiComponentListCard({ clusterName }: IYataiComponentListCardProps) {
    const { yataiComponentsInfo, queryKey } = useFetchYataiComponents(clusterName)
    const [isCreateYataiComponentOpen, setIsCreateYataiComponentOpen] = useState(false)
    const [wishToUpgradeType, setWishToUpgradeType] = useState<YataiComponentType>()
    const [wishToDeleteType, setWishToDeleteType] = useState<YataiComponentType>()
    const [wishToUpgradeTargetVersion, setWishToUpgradeTargetVersion] = useState<string>()
    const [t] = useTranslation()
    const { yataiComponentOperatorHelmChartsInfo } = useFetchYataiComponentOperatorHelmCharts()
    const [upgradeYataiComponentLoading, setUpgradeYataiComponentLoading] = useState(false)
    const [deleteYataiComponentLoading, setDeleteYataiComponentLoading] = useState(false)

    const handleCreateYataiComponent = useCallback(
        async (data: ICreateYataiComponentSchema) => {
            await createYataiComponent(clusterName, data)
            await yataiComponentsInfo.refetch()
            setIsCreateYataiComponentOpen(false)
        },
        [clusterName, yataiComponentsInfo]
    )

    const handleUpgradeYataiComponent = useCallback(
        async (type_: YataiComponentType) => {
            setUpgradeYataiComponentLoading(true)
            try {
                await handleCreateYataiComponent({
                    type: type_,
                })
                setWishToUpgradeType(undefined)
            } finally {
                setUpgradeYataiComponentLoading(false)
            }
        },
        [handleCreateYataiComponent]
    )

    const handleDeleteYataiComponent = useCallback(
        async (type: YataiComponentType) => {
            setDeleteYataiComponentLoading(true)
            try {
                await deleteYataiComponent(clusterName, type)
                await yataiComponentsInfo.refetch()
                setWishToDeleteType(undefined)
            } finally {
                setDeleteYataiComponentLoading(false)
            }
        },
        [clusterName, yataiComponentsInfo]
    )

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

    const [, theme] = useStyletron()

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
                columns={[t('type'), t('status'), 'Pods', t('version'), t('created_at'), t('operation')]}
                data={
                    yataiComponentsInfo.data?.map((yataiComponent) => {
                        const chartName = `yatai-${yataiComponent.type}-comp-operator`
                        const chart = yataiComponentOperatorHelmChartsInfo.data?.find(
                            (x) => x.metadata.name === chartName
                        )

                        return [
                            <Link
                                key={yataiComponent.type}
                                to={`/clusters/${clusterName}/yatai_components/${yataiComponent.type}`}
                            >
                                <YataiComponentTypeRender type={yataiComponent.type} />
                            </Link>,
                            <Tag
                                key={yataiComponent.type}
                                closeable={false}
                                variant={TagVariant.light}
                                kind={
                                    yataiComponent.release
                                        ? statusColorMap[yataiComponent.release.info.status]
                                        : undefined
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
                            <YataiComponentPodStatuses
                                key={yataiComponent.type}
                                clusterName={clusterName}
                                componentType={yataiComponent.type}
                            />,
                            yataiComponent.release ? yataiComponent.release.chart.metadata.version : '-',
                            yataiComponent.release ? formatDateTime(yataiComponent.release.info.last_deployed) : '-',
                            <div
                                key={yataiComponent.type}
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 10,
                                }}
                            >
                                {semver.lt(
                                    yataiComponent.release?.chart.metadata.version ?? '0.0.0',
                                    chart?.metadata.version ?? '0.0.0'
                                ) && (
                                    <Button
                                        size='compact'
                                        onClick={() => {
                                            setWishToUpgradeType(yataiComponent.type)
                                            setWishToUpgradeTargetVersion(chart?.metadata.version)
                                        }}
                                    >
                                        {t('upgrade to sth', [chart?.metadata.version])}
                                    </Button>
                                )}
                                <Button
                                    overrides={{
                                        Root: {
                                            style: {
                                                background: theme.colors.negative,
                                            },
                                        },
                                    }}
                                    size='compact'
                                    onClick={() => {
                                        setWishToDeleteType(yataiComponent.type)
                                    }}
                                >
                                    {t('delete')}
                                </Button>
                            </div>,
                        ]
                    }) ?? []
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
                    <YataiComponentForm onSubmit={handleCreateYataiComponent} clusterName={clusterName} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={wishToUpgradeType !== undefined}
                onClose={() => setWishToUpgradeType(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>
                    {t('do you want to upgrade yatai component sth to some version', [
                        wishToUpgradeType ? t(wishToUpgradeType) : '',
                        wishToUpgradeTargetVersion,
                    ])}
                </ModalHeader>
                <ModalFooter>
                    <ModalButton size='compact' kind='tertiary' onClick={() => setWishToUpgradeType(undefined)}>
                        {t('cancel')}
                    </ModalButton>
                    <ModalButton
                        size='compact'
                        onClick={(e) => {
                            e.preventDefault()
                            if (!wishToUpgradeType) {
                                return
                            }
                            handleUpgradeYataiComponent(wishToUpgradeType)
                        }}
                        isLoading={upgradeYataiComponentLoading}
                    >
                        {t('ok')}
                    </ModalButton>
                </ModalFooter>
            </Modal>
            <Modal
                isOpen={wishToDeleteType !== undefined}
                onClose={() => setWishToDeleteType(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>
                    {t('do you want to delete yatai component sth', [wishToDeleteType ? t(wishToDeleteType) : ''])}
                </ModalHeader>
                <ModalFooter>
                    <ModalButton size='compact' kind='tertiary' onClick={() => setWishToDeleteType(undefined)}>
                        {t('cancel')}
                    </ModalButton>
                    <ModalButton
                        size='compact'
                        overrides={{
                            BaseButton: {
                                style: {
                                    background: theme.colors.negative,
                                },
                            },
                        }}
                        onClick={(e) => {
                            e.preventDefault()
                            if (!wishToDeleteType) {
                                return
                            }
                            handleDeleteYataiComponent(wishToDeleteType)
                        }}
                        isLoading={deleteYataiComponentLoading}
                    >
                        {t('ok')}
                    </ModalButton>
                </ModalFooter>
            </Modal>
        </Card>
    )
}
