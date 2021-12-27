import { useBentoRepository, useBentoRepositoryLoading } from '@/hooks/useBentoRepository'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { fetchBentoRepository } from '@/services/bento_repository'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'

export default function BentoRepositorySidebar({ style }: IComposedSidebarProps) {
    const { bentoRepositoryName } = useParams<{ bentoRepositoryName: string }>()
    const bentoRepositoryInfo = useQuery(`fetchBentoRepository:${bentoRepositoryName}`, () =>
        fetchBentoRepository(bentoRepositoryName)
    )
    const { bentoRepository, setBentoRepository } = useBentoRepository()
    const { organization, setOrganization } = useOrganization()
    const { setBentoRepositoryLoading } = useBentoRepositoryLoading()
    useEffect(() => {
        setBentoRepositoryLoading(bentoRepositoryInfo.isLoading)
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
        setBentoRepositoryLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

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
        ],
        [bentoRepositoryName, t]
    )
    return (
        <BaseSidebar title={bentoRepositoryName} icon={resourceIconMapping.bento} navItems={navItems} style={style} />
    )
}
