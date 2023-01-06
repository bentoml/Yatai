import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { Skeleton } from 'baseui/skeleton'
import { useState } from 'react'
import PodStatuses from './PodStatuses'

interface IKubeResourcePodStatusesProps {
    clusterName: string
    resource: IKubeResourceSchema
    style?: React.CSSProperties
}

export default function KubeResourcePodStatuses({ clusterName, resource, style }: IKubeResourcePodStatusesProps) {
    const [pods, setPods] = useState<IKubePodSchema[]>([])
    const [podsLoading, setPodsLoading] = useState(true)

    useFetchClusterPods({
        clusterName,
        namespace: resource.namespace,
        selector: Object.keys(resource.match_labels)
            .reduce((p: string[], c: string) => {
                return [...p, `${c}=${resource.match_labels[c]}`]
            }, [])
            .join(','),
        setPods,
        setPodsLoading,
    })

    return (
        <div style={style}>
            {podsLoading ? (
                <Skeleton rows={1} animation />
            ) : (
                <PodStatuses showOperationIcons pods={pods ?? []} replicas={pods?.length ?? 0} />
            )}
        </div>
    )
}
