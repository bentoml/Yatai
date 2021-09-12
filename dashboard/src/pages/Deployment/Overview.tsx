import React, { useState } from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import { VscServerProcess } from 'react-icons/vsc'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import Card from '@/components/Card'
import { formatTime } from '@/utils/datetime'
import User from '@/components/User'
import PodsStatus from '@/components/PodsStatus'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchDeploymentPods } from '@/hooks/useFetchDeploymentPods'
import { useParams } from 'react-router-dom'
import { Skeleton } from 'baseui/skeleton'
import { StyledLink } from 'baseui/link'

export default function DeploymentOverview() {
    const { orgName, clusterName, deploymentName } =
        useParams<{ orgName: string; clusterName: string; deploymentName: string }>()
    const { deployment } = useDeployment()
    const { deploymentLoading } = useDeploymentLoading()
    const [pods, setPods] = useState<IKubePodSchema[]>()
    const [podsLoading, setPodsLoading] = useState(false)

    useFetchDeploymentPods({
        orgName,
        clusterName,
        deploymentName,
        setPods,
        setPodsLoading,
    })

    const [t] = useTranslation()

    return (
        <>
            <Card title={t('overview')} titleIcon={RiSurveyLine}>
                <Table
                    isLoading={deploymentLoading}
                    columns={[t('name'), 'URL', t('description'), t('creator'), t('created_at')]}
                    data={[
                        [
                            deployment?.name,
                            <div key={deployment?.uid}>
                                {deployment?.urls.map((url) => (
                                    <StyledLink key={url} href={url} target='_blank'>
                                        {url}
                                    </StyledLink>
                                ))}
                            </div>,
                            deployment?.description,
                            deployment?.creator && <User user={deployment?.creator} />,
                            deployment && formatTime(deployment.created_at),
                        ],
                    ]}
                />
            </Card>
            <Card title={t('replicas')} titleIcon={VscServerProcess}>
                {podsLoading ? <Skeleton rows={1} /> : <PodsStatus replicas={pods?.length ?? 0} pods={pods || []} />}
            </Card>
        </>
    )
}
