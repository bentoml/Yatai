import { useBento, useBentoLoading } from '@/hooks/useBento'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { fetchBento } from '@/services/bento'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import BaseSubLayout from './BaseSubLayout'

export interface IBentoLayoutProps {
    children: React.ReactNode
}

export default function BentoLayout({ children }: IBentoLayoutProps) {
    const { bentoName } = useParams<{ bentoName: string }>()
    const bentoInfo = useQuery(`fetchBento:${bentoName}`, () => fetchBento(bentoName))
    const { bento, setBento } = useBento()
    const { organization, setOrganization } = useOrganization()
    const { setBentoLoading } = useBentoLoading()
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
                title: t('sth list', [t('bento')]),
                path: '/bentos',
                icon: resourceIconMapping.bento,
            },
            {
                title: bentoName,
                path: `/bentos/${bentoName}`,
            },
        ],
        [bentoName, t]
    )

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/bentos/${bentoName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('version')]),
                path: `/bentos/${bentoName}/versions`,
                icon: resourceIconMapping.bento_version,
            },
        ],
        [bentoName, t]
    )

    return (
        <BaseSubLayout breadcrumbItems={breadcrumbItems} navItems={navItems}>
            {children}
        </BaseSubLayout>
    )
}
