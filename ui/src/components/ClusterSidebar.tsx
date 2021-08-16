import { useCluster, useClusterLoading } from '@/hooks/useCluster'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import { GrCluster } from 'react-icons/gr'
import { HiOutlineUserGroup } from 'react-icons/hi'
import { GoPackage } from 'react-icons/go'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { fetchCluster } from '@/services/cluster'

export default function ClusterSidebar({ style }: IComposedSidebarProps) {
    const { orgName, clusterName } = useParams<{ orgName: string; clusterName: string }>()
    const clusterInfo = useQuery(`fetchCluster:${orgName}:${clusterName}`, () => fetchCluster(orgName, clusterName))
    const { cluster, setCluster } = useCluster()
    const { setClusterLoading } = useClusterLoading()
    useEffect(() => {
        setClusterLoading(clusterInfo.isLoading)
        if (clusterInfo.isSuccess && clusterInfo.data.uid !== cluster?.uid) {
            setCluster(clusterInfo.data)
        } else if (clusterInfo.isLoading) {
            setCluster(undefined)
        }
    }, [cluster?.uid, clusterInfo.data, clusterInfo.isLoading, clusterInfo.isSuccess, setCluster, setClusterLoading])

    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/orgs/${orgName}/clusters/${clusterName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('bundle')]),
                path: `/orgs/${orgName}/clusters/${clusterName}/bundles`,
                icon: GoPackage,
            },
            {
                title: t('sth list', [t('member')]),
                path: `/orgs/${orgName}/clusters/${clusterName}/members`,
                icon: HiOutlineUserGroup,
            },
        ],
        [clusterName, orgName, t]
    )
    return <BaseSidebar title={clusterName} icon={<GrCluster />} navItems={navItems} style={style} />
}
