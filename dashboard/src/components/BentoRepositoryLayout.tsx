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
    const bentoInfo = useQuery(`fetchBentoRepository:${bentoRepositoryName}`, () =>
        fetchBentoRepository(bentoRepositoryName)
    )
    const { bentoRepository: bento, setBentoRepository: setBento } = useBentoRepository()
    const { organization, setOrganization } = useOrganization()
    const { setBentoRepositoryLoading: setBentoLoading } = useBentoRepositoryLoading()
    useEffect(() => {
        setBentoLoading(bentoInfo.isLoading)
        if (bentoInfo.isSuccess) {
            if (bentoInfo.data.uid !== bento?.uid) {
                setBento(bentoInfo.data)
            }
            if (bentoInfo.data.organization?.uid !== organization?.uid) {
                setOrganization(bentoInfo.data.organization)
            }
        } else if (bentoInfo.isLoading) {
            setBento(undefined)
        }
    }, [
        bento?.uid,
        bentoInfo.data,
        bentoInfo.isLoading,
        bentoInfo.isSuccess,
        organization?.uid,
        setBento,
        setBentoLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('sth list', [t('bento repository')]),
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
                title: t('sth list', [t('bento')]),
                path: `/bento_repositories/${bentoRepositoryName}/bentos`,
                icon: resourceIconMapping.bento,
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
