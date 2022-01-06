import { listOrganizationModelModules } from '@/services/organization'
import { useQuery } from 'react-query'

export function useFetchOrganizationModelModules() {
    return useQuery('fetchOrgModelModules', () => listOrganizationModelModules())
}
