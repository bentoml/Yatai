import { listYataiComponentOperatorHelmCharts } from '@/services/yatai_component'
import { useQuery } from 'react-query'

export function useFetchYataiComponentOperatorHelmCharts() {
    const queryKey = 'fetchYataiComponentOperatorHelmCharts'
    const yataiComponentOperatorHelmChartsInfo = useQuery(queryKey, listYataiComponentOperatorHelmCharts)
    return { yataiComponentOperatorHelmChartsInfo, queryKey }
}
