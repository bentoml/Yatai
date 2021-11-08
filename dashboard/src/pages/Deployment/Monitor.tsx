/* eslint-disable no-nested-ternary */
import Card from '@/components/Card'
import DeploymentMonitor from '@/components/DeploymentMonitor'
import { useDeployment } from '@/hooks/useDeployment'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import useTranslation from '@/hooks/useTranslation'
import { Skeleton } from 'baseui/skeleton'
import React from 'react'
import { AiOutlineDashboard } from 'react-icons/ai'
import { useParams } from 'react-router-dom'

export default function DeploymentMonitorPage() {
    const { clusterName } = useParams<{ clusterName: string }>()
    const { deployment } = useDeployment()
    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)

    const hasMonitoring = yataiComponentsInfo.data?.find((x) => x.type === 'monitoring') !== undefined
    const [t] = useTranslation()

    return (
        <Card title={t('monitor')} titleIcon={AiOutlineDashboard}>
            {hasMonitoring ? (
                deployment ? (
                    <DeploymentMonitor deployment={deployment} />
                ) : (
                    <Skeleton animation rows={3} />
                )
            ) : (
                t('please install yatai component first', [t('monitoring')])
            )}
        </Card>
    )
}
