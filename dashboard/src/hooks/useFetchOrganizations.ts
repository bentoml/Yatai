import { IListQuerySchema } from '@/schemas/list'
import { useQuery } from 'react-query'
import { listOrganizations } from '@/services/organization'
import qs from 'qs'

export function useFetchOrganizations(query: IListQuerySchema) {
    const organizationsInfo = useQuery(`fetchOrgs:${qs.stringify(query)}`, () => listOrganizations(query))
    return organizationsInfo
}
