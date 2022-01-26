import { useFetchBento } from '@/hooks/useFetchBento'
import { useBento, useBentoLoading } from '@/hooks/useBento'
import useTranslation from '@/hooks/useTranslation'
import React, { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'
import BaseSubLayout from './BaseSubLayout'

export interface IBentoLayoutProps {
    children: React.ReactNode
}

export default function BentoLayout({ children }: IBentoLayoutProps) {
    const { bentoRepositoryName, bentoVersion } = useParams<{ bentoRepositoryName: string; bentoVersion: string }>()
    const bentoInfo = useFetchBento(bentoRepositoryName, bentoVersion)
    const { setBento } = useBento()
    const { setBentoLoading } = useBentoLoading()
    useEffect(() => {
        setBento(bentoInfo.data)
        setBentoLoading(bentoInfo.isLoading)
    }, [bentoInfo, setBento, setBentoLoading])

    const [t] = useTranslation()

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('bentos'),
                path: '/bento_repositories',
                icon: resourceIconMapping.bento,
            },
            {
                title: `${bentoRepositoryName}:${bentoVersion}`,
                path: `/bento_repositories/${bentoRepositoryName}/bentos/${bentoVersion}`,
            },
        ],
        [bentoRepositoryName, bentoVersion, t]
    )

    return <BaseSubLayout breadcrumbItems={breadcrumbItems}>{children}</BaseSubLayout>
}
