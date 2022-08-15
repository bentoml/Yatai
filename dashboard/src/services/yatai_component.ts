import { IYataiComponentSchema } from '@/schemas/yatai_component'
import axios from 'axios'

export async function listYataiComponents(clusterName: string): Promise<IYataiComponentSchema[]> {
    const resp = await axios.get<IYataiComponentSchema[]>(`/api/v1/clusters/${clusterName}/yatai_components`)
    return resp.data
}

export async function listOrganizationYataiComponents(): Promise<IYataiComponentSchema[]> {
    const resp = await axios.get<IYataiComponentSchema[]>('/api/v1/yatai_components')
    return resp.data
}
