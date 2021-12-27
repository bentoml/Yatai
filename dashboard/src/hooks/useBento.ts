import useGlobalState from '@/hooks/global'

export const useBento = () => {
    const [bento, setBento] = useGlobalState('bento')

    return {
        bento,
        setBento,
    }
}

export const useBentoLoading = () => {
    const [bentoLoading, setBentoLoading] = useGlobalState('bentoLoading')

    return {
        bentoLoading,
        setBentoLoading,
    }
}
