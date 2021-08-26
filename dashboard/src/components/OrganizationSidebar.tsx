import { useOrganization, useOrganizationLoading } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { fetchOrganization } from '@/services/organization'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'

export default function OrganizationSidebar({ style }: IComposedSidebarProps) {
    const { orgName } = useParams<{ orgName: string }>()
    const orgInfo = useQuery(`fetchOrg:${orgName}`, () => fetchOrganization(orgName))
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
                path: `/orgs/${orgName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('bento')]),
                path: `/orgs/${orgName}/bentos`,
                icon: resourceIconMapping.bento,
            },
            {
                title: t('sth list', [t('cluster')]),
                path: `/orgs/${orgName}/clusters`,
                icon: resourceIconMapping.cluster,
            },
            {
                title: t('sth list', [t('member')]),
                path: `/orgs/${orgName}/members`,
                icon: resourceIconMapping.user_group,
            },
        ],
        [orgName, t]
    )
    return <BaseSidebar title={orgName} icon={resourceIconMapping.organization} navItems={navItems} style={style} />
}
