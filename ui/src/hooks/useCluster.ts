import useGlobalState from '@/hooks/global'

export const useCluster = () => {
    const [cluster, setCluster] = useGlobalState('cluster')

    return {
        cluster,
        setCluster,
    }
}

export const useClusterLoading = () => {
    const [clusterLoading, setClusterLoading] = useGlobalState('clusterLoading')

    return {
        clusterLoading,
        setClusterLoading,
    }
}
