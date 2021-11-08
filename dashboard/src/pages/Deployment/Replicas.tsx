import { VscServerProcess } from 'react-icons/vsc'
import Card from '@/components/Card'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchDeploymentPods } from '@/hooks/useFetchDeploymentPods'
import { useParams } from 'react-router-dom'
import PodList from '@/components/PodList'
import { useState } from 'react'
import useTranslation from '@/hooks/useTranslation'

export default function DeploymentReplicas() {
    const { clusterName, deploymentName } = useParams<{ clusterName: string; deploymentName: string }>()
    const [pods, setPods] = useState<IKubePodSchema[]>()
    const [podsLoading, setPodsLoading] = useState(false)
    const [t] = useTranslation()

    useFetchDeploymentPods({
        clusterName,
        deploymentName,
        setPods,
        setPodsLoading,
    })
    return (
        <Card title={t('replicas')} titleIcon={VscServerProcess}>
            <PodList loading={podsLoading} pods={pods ?? []} />
        </Card>
    )
}
