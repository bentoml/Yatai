import { useBentoRepository, useBentoRepositoryLoading } from '@/hooks/useBentoRepository'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { fetchBentoRepository } from '@/services/bento_repository'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import BaseSubLayout from './BaseSubLayout'

export interface IBentoRepositoryLayoutProps {
    children: React.ReactNode
}

export default function BentoRepositoryLayout({ children }: IBentoRepositoryLayoutProps) {
    const { bentoRepositoryName } = useParams<{ bentoRepositoryName: string }>()
    const bentoRepositoryInfo = useQuery(`fetchBentoRepository:${bentoRepositoryName}`, () =>
        fetchBentoRepository(bentoRepositoryName)
    )
    const { bentoRepository, setBentoRepository } = useBentoRepository()
    const { organization, setOrganization } = useOrganization()
    const { setBentoRepositoryLoading: setBentoLoading } = useBentoRepositoryLoading()
    useEffect(() => {
        setBentoLoading(bentoRepositoryInfo.isLoading)
        if (bentoRepositoryInfo.isSuccess) {
            if (bentoRepositoryInfo.data.uid !== bentoRepository?.uid) {
                setBentoRepository(bentoRepositoryInfo.data)
            }
            if (bentoRepositoryInfo.data.organization?.uid !== organization?.uid) {
                setOrganization(bentoRepositoryInfo.data.organization)
            }
        } else if (bentoRepositoryInfo.isLoading) {
            setBentoRepository(undefined)
        }
    }, [
        bentoRepository?.uid,
        bentoRepositoryInfo.data,
        bentoRepositoryInfo.isLoading,
        bentoRepositoryInfo.isSuccess,
        organization?.uid,
        setBentoRepository,
        setBentoLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('bento repositories'),
                path: '/bento_repositories',
                icon: resourceIconMapping.bento,
            },
            {
                title: bentoRepositoryName,
                path: `/bento_repositories/${bentoRepositoryName}`,
            },
        ],
        [bentoRepositoryName, t]
    )

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/bento_repositories/${bentoRepositoryName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('bentos'),
                path: `/bento_repositories/${bentoRepositoryName}/bentos`,
                icon: resourceIconMapping.bento,
            },
            {
                title: t('deployments'),
                path: `/bento_repositories/${bentoRepositoryName}/deployments`,
                icon: resourceIconMapping.deployment,
            },
        ],
        [bentoRepositoryName, t]
    )

    return (
        <BaseSubLayout breadcrumbItems={breadcrumbItems} navItems={navItems}>
            {children}
        </BaseSubLayout>
    )
}
