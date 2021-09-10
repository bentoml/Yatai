import React, { useEffect } from 'react'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { fetchCurrentUser } from '@/services/user'
import { useQuery } from 'react-query'
import { useFetchCurrentUserApiToken } from '@/hooks/useFetchCurrentUserApiToken'
import { IUserSchema } from '@/schemas/user'
import useTranslation from '@/hooks/useTranslation'

export interface IUserProfileCard {
    user: IUserSchema
}

const copyToClipboard = (text: string) => {
    if ('clipboard' in navigator) {
        return navigator.clipboard.writeText(text)
    }
    return document.execCommand('copy', true, text)
}

const ApiTokenSection = ({ apiToken }: { apiToken: string }) => {
    const [t] = useTranslation()
    const [isCopied, setIsCopied] = React.useState(false)
    const handleCopy = () => {
        copyToClipboard(apiToken)
        setIsCopied(true)
        setTimeout(() => {
            setIsCopied(false)
        }, 2000)
    }
    return (
        <div>
            <span style={{ paddingRight: '5px' }}>
                {t('api token')}: {apiToken}
            </span>
            <button onClick={handleCopy} type='button'>
                <span>{isCopied ? 'Copied' : 'Copy'}</span>
            </button>
        </div>
    )
}

const UserProfileCard = ({ user }: IUserProfileCard) => {
    const [t] = useTranslation()
    return (
        <div>
            <h2>{t('user profile')}</h2>
            <h3>{user.name}</h3>
            <ApiTokenSection apiToken={user.api_token} />
        </div>
    )
}

export default function UserProfile() {
    const { currentUser, setCurrentUser } = useCurrentUser()
    const userInfo = useQuery('currentUser', fetchCurrentUser)
    useEffect(() => {
        if (userInfo.isSuccess) {
            setCurrentUser(userInfo.data)
        }
    }, [userInfo.data, userInfo.isSuccess, setCurrentUser])
    useFetchCurrentUserApiToken()

    if (!currentUser) {
        return <div>Loading...</div>
    }

    return <UserProfileCard user={currentUser} />
}
