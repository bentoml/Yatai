import { IDeploymentSnapshotSchema } from './deployment_snapshot'
import { IKubeEventSchema } from './kube_event'

export type PodStatusPhase = 'Running' | 'Unknown' | 'ContainerCreating' | 'Pending' | 'Terminating'

export interface IPodStatusSchema {
    phase: PodStatusPhase
    ready: boolean
    start_time: string
    is_old: boolean
    is_canary: boolean
    host_ip: string
}

export type KubePodStatusPhase = 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown' | 'Terminating'

export interface IKubePodStatusSchema {
    status: KubePodStatusPhase
}

export interface IKubePodSchema {
    name: string
    node_name: string
    image: string
    commit_id: string
    status: IPodStatusSchema
    pod_status: IKubePodStatusSchema
    warnings?: IKubeEventSchema[]
    deployment_snapshot?: IDeploymentSnapshotSchema
    raw_status?: {
        podIP?: string
        containerStatuses?: {
            name: string
            image: string
            started: boolean
            ready: boolean
            lastState: {
                terminated?: {
                    reason: string
                    message: string
                    startedAt: string
                    finishedAt: string
                    exitCode: number
                }
            }
            state: {
                waiting?: {
                    message: string
                    reason: string
                }
                running?: {
                    startedAt: string
                }
                terminated?: {
                    reason: 'Completed' | 'Running' | 'unknown'
                    message: string
                    startedAt: string
                    finishedAt: string
                    exitCode: number
                }
            }
        }[]
    }
}
