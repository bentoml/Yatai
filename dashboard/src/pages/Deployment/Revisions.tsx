import React from 'react'
import { useParams } from 'react-router-dom'
import DeploymentRevisionListCard from '@/components/DeploymentRevisionListCard'
import KubePodEvents from '@/components/KubePodEvents'
import Card from '@/components/Card'
import useTranslation from '@/hooks/useTranslation'
import { MdEventNote } from 'react-icons/md'

export default function DeploymentRevisions() {
    const { clusterName, deploymentName } = useParams<{ clusterName: string; deploymentName: string }>()

    const [t] = useTranslation()

    return (
        <>
            <DeploymentRevisionListCard clusterName={clusterName} deploymentName={deploymentName} />
            <Card title={t('events')} titleIcon={MdEventNote}>
                <KubePodEvents
                    open
                    width='auto'
                    height={200}
                    clusterName={clusterName}
                    deploymentName={deploymentName}
                />
            </Card>
        </>
    )
}
