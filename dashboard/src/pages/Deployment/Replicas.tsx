import { VscServerProcess } from 'react-icons/vsc'
import Card from '@/components/Card'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchDeploymentPods } from '@/hooks/useFetchDeploymentPods'
import { useParams } from 'react-router-dom'
import PodList from '@/components/PodList'
import { useState } from 'react'
import useTranslation from '@/hooks/useTranslation'
import KubePodEvents from '@/components/KubePodEvents'
import { MdEventNote } from 'react-icons/md'
import { useDeployment } from '@/hooks/useDeployment'

export default function DeploymentReplicas() {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const [pods, setPods] = useState<IKubePodSchema[]>()
    const [podsLoading, setPodsLoading] = useState(false)
    const [t] = useTranslation()
    const { deployment } = useDeployment()

    useFetchDeploymentPods({
        clusterName,
        kubeNamespace,
        deploymentName,
        setPods,
        setPodsLoading,
    })

    return (
        <div>
            <Card title={t('replicas')} titleIcon={VscServerProcess}>
                <PodList deployment={deployment} loading={podsLoading} pods={pods ?? []} groupByRunner />
            </Card>
            <Card title={t('events')} titleIcon={MdEventNote}>
                <KubePodEvents
                    open
                    width='auto'
                    height={200}
                    clusterName={clusterName}
                    namespace={kubeNamespace}
                    deploymentName={deploymentName}
                />
            </Card>
        </div>
    )
}
