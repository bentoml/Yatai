import { IInfoSchema } from '@/schemas/info'
import axios from 'axios'
import { useQuery } from 'react-query'

export function useFetchInfo() {
    const infoInfo = useQuery('fetchInfo', async (): Promise<IInfoSchema> => {
        const resp = await axios.get<IInfoSchema>('/api/v1/info')
        return resp.data
    })
    return infoInfo
}
