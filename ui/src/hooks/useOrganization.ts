import useGlobalState from '@/hooks/global'

export const useOrganization = () => {
    const [organization, setOrganization] = useGlobalState('organization')

    return {
        organization,
        setOrganization,
    }
}

export const useOrganizationLoading = () => {
    const [organizationLoading, setOrganizationLoading] = useGlobalState('organizationLoading')

    return {
        organizationLoading,
        setOrganizationLoading,
    }
}
