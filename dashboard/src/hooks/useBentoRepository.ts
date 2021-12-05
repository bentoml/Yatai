import useGlobalState from '@/hooks/global'

export const useBentoRepository = () => {
    const [bentoRepository, setBentoRepository] = useGlobalState('bentoRepository')

    return {
        bentoRepository,
        setBentoRepository,
    }
}

export const useBentoRepositoryLoading = () => {
    const [bentoRepositoryLoading, setBentoRepositoryLoading] = useGlobalState('bentoRepositoryLoading')

    return {
        bentoRepositoryLoading,
        setBentoRepositoryLoading,
    }
}
