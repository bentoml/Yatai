import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useState } from 'react'
import PodList from './PodList'

interface IPodsProps {
    clusterName: string
    namespace: string
    selector: string
}

export default function Pods({ clusterName, namespace, selector }: IPodsProps) {
    const [pods, setPods] = useState<IKubePodSchema[]>([])
    const [podsLoading, setPodsLoading] = useState(false)
    useFetchClusterPods({
        clusterName,
        namespace,
        selector,
        setPods,
        setPodsLoading,
    })

    return <PodList clusterName={clusterName} loading={podsLoading} pods={pods} />
}
