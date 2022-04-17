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
import { GiAbstract006, GiAbstract045 } from 'react-icons/gi'

export default function DeploymentReplicas() {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const [pods, setPods] = useState<IKubePodSchema[]>()
    const [podsLoading, setPodsLoading] = useState(false)
    const [t] = useTranslation()

    useFetchDeploymentPods({
        clusterName,
        kubeNamespace,
        deploymentName,
        setPods,
        setPodsLoading,
    })

    const apiServerPods = pods?.filter((pod) => !pod.runner_name) ?? []

    const runnerPodsGroup =
        pods?.reduce((acc, pod) => {
            const { runner_name: runnerName } = pod
            if (!runnerName) {
                return acc
            }
            const pods_ = acc[runnerName] ?? []
            return {
                ...acc,
                [runnerName]: [...pods_, pod],
            }
        }, {} as Record<string, IKubePodSchema[]>) ?? {}

    const runnerNames = Object.keys(runnerPodsGroup).sort((a, b) => {
        return runnerPodsGroup[a][0].name.localeCompare(runnerPodsGroup[b][0].name)
    })

    return (
        <div>
            <Card title={t('replicas')} titleIcon={VscServerProcess}>
                <Card title='Api Server' titleIcon={GiAbstract006}>
                    <PodList loading={podsLoading} pods={apiServerPods} />
                </Card>
                {runnerNames.map((runnerName) => (
                    <Card
                        key={runnerName}
                        title={
                            <span>
                                Runner <span style={{ fontWeight: 'bolder' }}>{runnerName}</span>
                            </span>
                        }
                        titleIcon={GiAbstract045}
                    >
                        <PodList loading={podsLoading} pods={runnerPodsGroup[runnerName]} />
                    </Card>
                ))}
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
