import { listYataiComponents, listOrganizationYataiComponents } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponents(clusterName: string) {
    const queryKey = `fetchYataiComponents:${clusterName}`
    const yataiComponentsInfo = useQuery(queryKey, () => listYataiComponents(clusterName))
    return { yataiComponentsInfo }
}

export function useFetchOrganizationYataiComponents() {
    const queryKey = 'fetchOrganizationYataiComponents'
    const yataiComponentsInfo = useQuery(queryKey, listOrganizationYataiComponents)
    return { yataiComponentsInfo }
}
