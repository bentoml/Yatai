import useGlobalState from '@/hooks/global'

export const useModelRepository = () => {
    const [modelRepository, setModelRepository] = useGlobalState('modelRepository')

    return {
        modelRepository,
        setModelRepository,
    }
}

export const useModelRepositoryLoading = () => {
    const [modelRepositoryLoading, setModelRepositoryLoading] = useGlobalState('modelRepositoryLoading')

    return {
        modelRepositoryLoading,
        setModelRepositoryLoading,
    }
}
