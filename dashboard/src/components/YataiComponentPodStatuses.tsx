import { useFetchYataiComponentHelmChartReleaseResources } from '@/hooks/useFetchYataiComponentHelmChartReleaseResources'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { YataiComponentType } from '@/schemas/yatai_component'
import { Skeleton } from 'baseui/skeleton'
import { useState } from 'react'
import KubeResourcePodStatuses from './KubeResourcePodStatuses'

interface IYataiComponentPodStatusesProps {
    clusterName: string
    componentType: YataiComponentType
}

export function YataiComponentPodStatuses({ clusterName, componentType }: IYataiComponentPodStatusesProps) {
    const [kubeResources, setKubeResources] = useState<IKubeResourceSchema[]>([])
    const [kubeResourcesLoading, setKubeResourcesLoading] = useState(true)

    useFetchYataiComponentHelmChartReleaseResources(
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
                    <KubeResourcePodStatuses key={idx} clusterName={clusterName} resource={x} />
                ))
            )}
        </div>
    )
}
