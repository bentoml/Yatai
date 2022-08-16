import { listOrganizationMembers } from '@/services/organization_member'
import { useQuery } from 'react-query'
import { useOrganization } from './useOrganization'

export function useFetchOrganizationMembers() {
    const { organization } = useOrganization()
    return useQuery(`fetchOrgMembers:${organization?.name}`, () => listOrganizationMembers())
}
