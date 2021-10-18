import { useFetchYataiComponentHelmChartReleaseResources } from '@/hooks/useFetchYataiComponentHelmChartReleaseResources'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { YataiComponentType } from '@/schemas/yatai_component'
import { Skeleton } from 'baseui/skeleton'
import { useState } from 'react'
import KubeResourcePodStatuses from './KubeResourcePodStatuses'

interface IYataiComponentPodStatusesProps {
    orgName: string
    clusterName: string
    componentType: YataiComponentType
}

export function YataiComponentPodStatuses({ orgName, clusterName, componentType }: IYataiComponentPodStatusesProps) {
    const [kubeResources, setKubeResources] = useState<IKubeResourceSchema[]>([])
    const [kubeResourcesLoading, setKubeResourcesLoading] = useState(true)

    useFetchYataiComponentHelmChartReleaseResources(
        orgName,
        clusterName,
        componentType,
        setKubeResources,
        setKubeResourcesLoading
    )

    return (
        <div
            style={{
                display: 'flex',
            }}
        >
            {kubeResourcesLoading ? (
                <Skeleton rows={3} animation />
            ) : (
                kubeResources.map((x, idx) => (
                    <KubeResourcePodStatuses key={idx} orgName={orgName} clusterName={clusterName} resource={x} />
                ))
            )}
        </div>
    )
}
