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
import Pods from '@/components/Pods'
import { GrDocker } from 'react-icons/gr'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import _ from 'lodash'

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

    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)
    let imageBuilderPodNamespace = 'yatai-builders'
    let imageBuilderPodSelector = `yatai.ai/bento-repository=${deployment?.latest_revision?.targets[0]?.bento.repository.name},yatai.ai/bento=${deployment?.latest_revision?.targets[0]?.bento.version}`
    if (
        _.startsWith(
            yataiComponentsInfo?.data?.find((component) => component.name === 'deployment')?.manifest
                ?.latest_crd_version ?? 'v1alpha2',
            'v2'
        )
    ) {
        imageBuilderPodNamespace = kubeNamespace
        imageBuilderPodSelector = `yatai.ai/is-bento-image-builder=true,yatai.ai/bento-repository=${deployment?.latest_revision?.targets[0]?.bento.repository.name},yatai.ai/bento=${deployment?.latest_revision?.targets[0]?.bento.version}`
    }

    return (
        <div>
            <Card title={t('replicas')} titleIcon={VscServerProcess}>
                <PodList deployment={deployment} loading={podsLoading} pods={pods ?? []} groupByRunner />
            </Card>
            {deployment?.latest_revision?.targets[0] && (
                <Card title={t('docker image builder pods')} titleIcon={GrDocker}>
                    <Pods
                        clusterName={clusterName}
                        namespace={imageBuilderPodNamespace}
                        selector={imageBuilderPodSelector}
                    />
                </Card>
            )}
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
