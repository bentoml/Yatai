import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listOrganizations } from '@/services/organization'

export function useFetchOrganizations(query: IListQuerySchema) {
    const organizationsInfo = useQuery('fetchOrgs', () => listOrganizations(query))
    return organizationsInfo
}
