import { useOrganization, useOrganizationLoading } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { fetchOrganization } from '@/services/organization'
import { RiSurveyLine } from 'react-icons/ri'
import { GrServerCluster, GrOrganization } from 'react-icons/gr'
import { HiOutlineUserGroup } from 'react-icons/hi'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from './BaseSidebar'

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
                icon: <RiSurveyLine />,
            },
            {
                title: t('sth list', [t('cluster')]),
                path: `/orgs/${orgName}/clusters`,
                icon: <GrServerCluster />,
            },
            {
                title: t('sth list', [t('member')]),
                path: `/orgs/${orgName}/members`,
                icon: <HiOutlineUserGroup />,
            },
        ],
        [orgName, t]
    )
    return <BaseSidebar title={orgName} icon={<GrOrganization />} navItems={navItems} style={style} />
}
