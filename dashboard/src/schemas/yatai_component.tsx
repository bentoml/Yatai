/* eslint-disable @typescript-eslint/no-explicit-any */
import { IHelmChart } from './helm_chart'

export type YataiComponentType = 'deployment' | 'logging' | 'monitoring'
export type YataiComponentReleaseStatus =
    | 'unknown'
    | 'deployed'
    | 'uninstalled'
    | 'superseded'
    | 'failed'
    | 'uninstalling'
    | 'pending-install'
    | 'pending-upgrade'
    | 'pending-rollback'

export interface IYataiComponentSchema {
    type: YataiComponentType
    release?: {
        name: string
        info: {
            first_deployed: string
            last_deployed: string
            deleted: string
            description: string
            status: YataiComponentReleaseStatus
            notes: string
        }
        chart: IHelmChart
        config: Record<string, any>
        version: number
        namespace: string
        labels: Record<string, string>
    }
}

export interface ICreateYataiComponentSchema {
    type: YataiComponentType
}
