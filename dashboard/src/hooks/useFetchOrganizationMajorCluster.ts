import { fetchOrganizationMajorCluster } from '@/services/organization'
import { useQuery } from 'react-query'

export function useFetchOrganizationMajorCluster() {
    return useQuery('fetchOrganizationMajorCluster', fetchOrganizationMajorCluster)
}
