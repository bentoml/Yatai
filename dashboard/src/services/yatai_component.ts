import axios from 'axios'
import { ICreateYataiComponentSchema, IYataiComponentSchema, YataiComponentType } from '@/schemas/yatai_component'
import { IHelmChart } from '@/schemas/helm_chart'

export async function listClusterYataiComponents(clusterName: string): Promise<IYataiComponentSchema[]> {
    const resp = await axios.get<IYataiComponentSchema[]>(`/api/v1/clusters/${clusterName}/yatai_components`)
    return resp.data
}

export async function listYataiComponentOperatorHelmCharts(): Promise<IHelmChart[]> {
    const resp = await axios.get<IHelmChart[]>('/api/v1/yatai_component_operator_helm_charts')
    return resp.data
}

export async function fetchYataiComponent(
    clusterName: string,
    componentType: YataiComponentType
): Promise<IYataiComponentSchema> {
    const resp = await axios.get<IYataiComponentSchema>(
        `/api/v1/clusters/${clusterName}/yatai_components/${componentType}`
    )
    return resp.data
}

export async function createYataiComponent(
    clusterName: string,
    data: ICreateYataiComponentSchema
): Promise<IYataiComponentSchema> {
    const resp = await axios.post<IYataiComponentSchema>(`/api/v1/clusters/${clusterName}/yatai_components`, data)
    return resp.data
}

export async function deleteYataiComponent(
    clusterName: string,
    componentType: YataiComponentType
): Promise<IYataiComponentSchema> {
    const resp = await axios.delete<IYataiComponentSchema>(
        `/api/v1/clusters/${clusterName}/yatai_components/${componentType}`
    )
    return resp.data
}
