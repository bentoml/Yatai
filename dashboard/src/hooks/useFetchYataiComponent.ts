import { YataiComponentType } from '@/schemas/yatai_component'
import { fetchYataiComponent } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponent(clusterName: string, componentType: YataiComponentType) {
    const queryKey = `fetchYataiComponent:${clusterName}:${componentType}`
    const yataiComponentInfo = useQuery(queryKey, () => fetchYataiComponent(clusterName, componentType))
    return { yataiComponentInfo, queryKey }
}
