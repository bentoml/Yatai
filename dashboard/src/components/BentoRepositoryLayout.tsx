import { useBentoRepository, useBentoRepositoryLoading } from '@/hooks/useBentoRepository'
import useTranslation from '@/hooks/useTranslation'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useHistory, useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { fetchBentoRepository, listBentoRepositoryDeployments } from '@/services/bento_repository'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import { Button } from 'baseui/button'
import qs from 'qs'
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

    const deploymentsInfo = useQuery(`bentoRepository:${bentoRepositoryName}:deployments`, () =>
        listBentoRepositoryDeployments(bentoRepositoryName, { count: 0, start: 0 })
    )

    const history = useHistory()

    return (
        <BaseSubLayout
            extra={
                <Button
                    isLoading={deploymentsInfo.isLoading}
                    kind='tertiary'
                    size='mini'
                    onClick={() =>
                        history.push(
                            `/deployments?${qs.stringify({
                                q: `bento_repository:${bentoRepositoryName}`,
                            })}`
                        )
                    }
                >
                    {t('n deployments', [deploymentsInfo.data?.total ?? '-'])}
                </Button>
            }
            breadcrumbItems={breadcrumbItems}
        >
            {children}
        </BaseSubLayout>
    )
}
