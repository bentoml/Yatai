import useGlobalState from '@/hooks/global'

export const useCurrentUser = () => {
    const [currentUser, setCurrentUser] = useGlobalState('currentUser')

    return {
        currentUser,
        setCurrentUser,
    }
}
