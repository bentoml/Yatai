import { listOrganizationModelModules } from '@/services/organization'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchOrganizationModelModules() {
    const { organization } = useOrganization()
    return useQuery(`fetchOrgModelModules:${organization}`, () => listOrganizationModelModules())
}
