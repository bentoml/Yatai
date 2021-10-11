import { IYataiComponentSchema } from '@/schemas/yatai_component'
import { listClusterYataiComponents } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponents(orgName?: string, clusterName?: string) {
    const queryKey = `fetchYataiComponents:${orgName}:${clusterName}`
    const yataiComponentsInfo = useQuery(queryKey, () =>
        orgName && clusterName
            ? listClusterYataiComponents(orgName, clusterName)
            : (new Promise((resolve) => {
                  resolve([] as IYataiComponentSchema[])
              }) as Promise<IYataiComponentSchema[]>)
    )
    return { yataiComponentsInfo, queryKey }
}
