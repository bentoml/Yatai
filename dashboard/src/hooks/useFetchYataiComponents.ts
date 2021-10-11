import { listClusterYataiComponents } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponents(orgName: string, clusterName: string) {
    const queryKey = `fetchYataiComponents:${orgName}:${clusterName}`
    const yataiComponentsInfo = useQuery(queryKey, () => listClusterYataiComponents(orgName, clusterName))
    return { yataiComponentsInfo, queryKey }
}
