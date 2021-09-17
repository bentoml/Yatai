import React from 'react'
import { useParams } from 'react-router-dom'
import DeploymentSnapshotListCard from '@/components/DeploymentSnapshotListCard'
import DeploymentKubeEvents from '@/components/DeploymentKubeEvents'
import Card from '@/components/Card'
import useTranslation from '@/hooks/useTranslation'
import { MdEventNote } from 'react-icons/md'

export default function DeploymentSnapshots() {
    const { orgName, clusterName, deploymentName } =
        useParams<{ orgName: string; clusterName: string; deploymentName: string }>()

    const [t] = useTranslation()

    return (
        <>
            <DeploymentSnapshotListCard orgName={orgName} clusterName={clusterName} deploymentName={deploymentName} />
            <Card title={t('events')} titleIcon={MdEventNote}>
                <DeploymentKubeEvents
                    open
                    width='auto'
                    orgName={orgName}
                    clusterName={clusterName}
                    deploymentName={deploymentName}
                />
            </Card>
        </>
    )
}
