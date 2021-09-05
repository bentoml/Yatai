import useTranslation from '@/hooks/useTranslation'
import Home from '@/pages/Yatai/Home'
import { useMemo } from 'react'

export interface IRouteItem {
    title: string
    path?: string
    icon?: string
    canBeMenuItem?: boolean
    requiredAdmin?: boolean
    items?: IRouteItem[]
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    component?: React.ComponentType<any>
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    layout?: React.ComponentType<any>
}

export default function useRouteItems() {
    const [t] = useTranslation()

    const rootRouteItem: IRouteItem = useMemo(
        () => ({
            title: t('homepage'),
            path: '/',
            items: [],
            component: Home,
        }),
        [t]
    )

    return rootRouteItem
}
