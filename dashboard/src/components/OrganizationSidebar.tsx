import { useOrganization, useOrganizationLoading } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { fetchOrganization } from '@/services/organization'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'
import { FiActivity } from 'react-icons/fi'

export default function OrganizationSidebar({ style }: IComposedSidebarProps) {
    const orgInfo = useQuery('fetchOrg', () => fetchOrganization())
    const { organization, setOrganization } = useOrganization()
    const { setOrganizationLoading } = useOrganizationLoading()
    useEffect(() => {
        setOrganizationLoading(orgInfo.isLoading)
        if (orgInfo.isSuccess && orgInfo.data.uid !== organization?.uid) {
            setOrganization(orgInfo.data)
        } else if (orgInfo.isLoading) {
            setOrganization(undefined)
        }
    }, [orgInfo.data, orgInfo.isLoading, orgInfo.isSuccess, organization?.uid, setOrganization, setOrganizationLoading])

    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: '/',
                icon: RiSurveyLine,
            },
            {
                title: t('models'),
                path: '/models',
                icon: resourceIconMapping.model,
                activePathPattern: /^\/(models|model_repositories)\/?/,
            },
            {
                title: t('bentos'),
                path: '/bento_repositories',
                icon: resourceIconMapping.bento,
            },
            {
                title: t('deployments'),
                path: '/deployments',
                icon: resourceIconMapping.deployment,
                activePathPattern: /^\/(deployments|clusters\/[^/]+\/deployments\/[^/]+)\/?/,
            },
            {
                title: t('events'),
                path: '/events',
                icon: FiActivity,
            },
            {
                title: t('clusters'),
                path: '/clusters',
                icon: resourceIconMapping.cluster,
            },
        ],
        [t]
    )
    return <BaseSidebar navItems={navItems} style={style} />
}
