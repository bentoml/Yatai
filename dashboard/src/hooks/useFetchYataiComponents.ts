import { IYataiComponentSchema } from '@/schemas/yatai_component'
import { listClusterYataiComponents } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponents(clusterName?: string) {
    const queryKey = `fetchYataiComponents:${clusterName}`
    const yataiComponentsInfo = useQuery(queryKey, () =>
        clusterName
            ? listClusterYataiComponents(clusterName)
            : (new Promise((resolve) => {
                  resolve([] as IYataiComponentSchema[])
              }) as Promise<IYataiComponentSchema[]>)
    )
    return { yataiComponentsInfo, queryKey }
}
