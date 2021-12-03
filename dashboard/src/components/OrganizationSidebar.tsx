import { useOrganization, useOrganizationLoading } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { fetchOrganization } from '@/services/organization'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'

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
                title: t('sth list', [t('model')]),
                path: '/model_versions',
                icon: resourceIconMapping.model,
            },
            {
                title: t('sth list', [t('bento')]),
                path: '/bento_versions',
                icon: resourceIconMapping.bento,
            },
            {
                title: t('sth list', [t('deployment')]),
                path: '/deployments',
                icon: resourceIconMapping.deployment,
            },
            {
                title: t('sth list', [t('cluster')]),
                path: '/clusters',
                icon: resourceIconMapping.cluster,
            },
            {
                title: t('sth list', [t('member')]),
                path: '/members',
                icon: resourceIconMapping.user_group,
            },
        ],
        [t]
    )
    return <BaseSidebar navItems={navItems} style={style} settingsPath='/settings' />
}
