import { IDeploymentTargetSchema } from './deployment_target'
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

export interface IContainerState {
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

export interface IContainerStatus {
    name: string
    image: string
    imageID: string
    started: boolean
    ready: boolean
    lastState: IContainerState
    state: IContainerState
    restartCount: number
    containerID: string
}

export interface IKubePodSchema {
    name: string
    namespace: string
    annotations?: Record<string, string>
    labels?: Record<string, string>
    node_name: string
    runner_name?: string
    image: string
    commit_id: string
    status: IPodStatusSchema
    pod_status: IKubePodStatusSchema
    warnings?: IKubeEventSchema[]
    deployment_target?: IDeploymentTargetSchema
    raw_status?: {
        podIP: string
        startTime?: string
        initContainerStatuses?: IContainerStatus[]
        containerStatuses?: IContainerStatus[]
    }
}
