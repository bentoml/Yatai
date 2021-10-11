import axios from 'axios'
import { ICreateYataiComponentSchema, IYataiComponentSchema, YataiComponentType } from '@/schemas/yatai_component'

export async function listClusterYataiComponents(
    orgName: string,
    clusterName: string
): Promise<IYataiComponentSchema[]> {
    const resp = await axios.get<IYataiComponentSchema[]>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/yatai_components`
    )
    return resp.data
}

export async function fetchYataiComponent(
    orgName: string,
    clusterName: string,
    componentType: YataiComponentType
): Promise<IYataiComponentSchema> {
    const resp = await axios.get<IYataiComponentSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/yatai_components/${componentType}`
    )
    return resp.data
}

export async function createYataiComponent(
    orgName: string,
    clusterName: string,
    data: ICreateYataiComponentSchema
): Promise<IYataiComponentSchema> {
    const resp = await axios.post<IYataiComponentSchema>(
        `/api/v1/orgs/${orgName}/clusters/${clusterName}/yatai_components`,
        data
    )
    return resp.data
}
