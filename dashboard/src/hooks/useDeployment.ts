import useGlobalState from '@/hooks/global'

export const useDeployment = () => {
    const [deployment, setDeployment] = useGlobalState('deployment')

    return {
        deployment,
        setDeployment,
    }
}

export const useDeploymentLoading = () => {
    const [deploymentLoading, setDeploymentLoading] = useGlobalState('deploymentLoading')

    return {
        deploymentLoading,
        setDeploymentLoading,
    }
}
