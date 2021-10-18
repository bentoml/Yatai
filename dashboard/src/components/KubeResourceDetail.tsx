import { useFetchClusterPods } from '@/hooks/useFetchClusterPods'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { StatefulTooltip } from 'baseui/tooltip'
import { useState } from 'react'
import { AiOutlineDeploymentUnit } from 'react-icons/ai'
import { GiMatterStates, GiDaemonPull } from 'react-icons/gi'
import PodList from './PodList'

interface IKubeResourceDetailProps {
    orgName: string
    clusterName: string
    resource: IKubeResourceSchema
    style?: React.CSSProperties
}

export default function KubeResourceDetail({ orgName, clusterName, resource, style }: IKubeResourceDetailProps) {
    const [pods, setPods] = useState<IKubePodSchema[]>([])
    const [podsLoading, setPodsLoading] = useState(false)

    useFetchClusterPods({
        orgName,
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
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                    padding: '10px 0',
                }}
            >
                <StatefulTooltip content={resource.kind} showArrow>
                    <div
                        style={{
                            cursor: 'pointer',
                        }}
                    >
                        {resource.kind === 'Deployment' && <AiOutlineDeploymentUnit />}
                        {resource.kind === 'StatefulSet' && <GiMatterStates />}
                        {resource.kind === 'DaemonSet' && <GiDaemonPull />}
                    </div>
                </StatefulTooltip>
                <div
                    style={{
                        fontWeight: 600,
                    }}
                >
                    {resource.name}
                </div>
            </div>
            <div>
                <PodList loading={podsLoading} pods={pods} />
            </div>
        </div>
    )
}
