import { listOrganizationMembers } from '@/services/organization_member'
import { useQuery } from 'react-query'

export function useFetchOrganizationMembers() {
    return useQuery('fetchOrgMembers', () => listOrganizationMembers())
}
