import { YataiComponentType } from '@/schemas/yatai_component'
import { fetchYataiComponent } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponent(orgName: string, clusterName: string, componentType: YataiComponentType) {
    const queryKey = `fetchYataiComponent:${orgName}:${clusterName}:${componentType}`
    const yataiComponentInfo = useQuery(queryKey, () => fetchYataiComponent(orgName, clusterName, componentType))
    return { yataiComponentInfo, queryKey }
}
