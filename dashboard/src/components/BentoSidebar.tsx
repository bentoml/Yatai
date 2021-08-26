import { useBento, useBentoLoading } from '@/hooks/useBento'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { fetchBento } from '@/services/bento'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'

export default function BentoSidebar({ style }: IComposedSidebarProps) {
    const { orgName, bentoName } = useParams<{ orgName: string; bentoName: string }>()
    const bentoInfo = useQuery(`fetchBento:${orgName}:${bentoName}`, () => fetchBento(orgName, bentoName))
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

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/orgs/${orgName}/bentos/${bentoName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('version')]),
                path: `/orgs/${orgName}/bentos/${bentoName}/versions`,
                icon: resourceIconMapping.bento_version,
            },
        ],
        [bentoName, orgName, t]
    )
    return <BaseSidebar title={bentoName} icon={resourceIconMapping.bento} navItems={navItems} style={style} />
}
