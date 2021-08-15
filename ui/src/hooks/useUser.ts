import useGlobalState from '@/hooks/global'

export const useUser = () => {
    const [user, setUser] = useGlobalState('user')

    return {
        user,
        setUser,
    }
}

export const useUserLoading = () => {
    const [userLoading, setUserLoading] = useGlobalState('userLoading')

    return {
        userLoading,
        setUserLoading,
    }
}
