import useGlobalState from '@/hooks/global'

export const useBundle = () => {
    const [bundle, setBundle] = useGlobalState('bundle')

    return {
        bundle,
        setBundle,
    }
}

export const useBundleLoading = () => {
    const [bundleLoading, setBundleLoading] = useGlobalState('bundleLoading')

    return {
        bundleLoading,
        setBundleLoading,
    }
}
