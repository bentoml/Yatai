import useGlobalState from '@/hooks/global'

export const useModel = () => {
    const [model, setModel] = useGlobalState('model')

    return {
        model,
        setModel,
    }
}

export const useModelLoading = () => {
    const [modelLoading, setModelLoading] = useGlobalState('modelLoading')

    return {
        modelLoading,
        setModelLoading,
    }
}
