import { listYataiComponents, listOrganizationYataiComponents } from '@/services/yatai_component'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchYataiComponents(clusterName: string) {
    const { organization } = useOrganization()
    const queryKey = `fetchYataiComponents:${organization?.name}:${clusterName}`
    const yataiComponentsInfo = useQuery(queryKey, () => listYataiComponents(clusterName))
    return { yataiComponentsInfo }
}

export function useFetchOrganizationYataiComponents() {
    const { organization } = useOrganization()
    const queryKey = `fetchOrganizationYataiComponents:${organization?.name}`
    const yataiComponentsInfo = useQuery(queryKey, listOrganizationYataiComponents)
    return { yataiComponentsInfo }
}
