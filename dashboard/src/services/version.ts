import { IVersionSchema } from '@/schemas/version'
import axios from 'axios'

export async function fetchVersion(): Promise<IVersionSchema> {
    const resp = await axios.get<IVersionSchema>('/api/v1/version')
    return resp.data
}
