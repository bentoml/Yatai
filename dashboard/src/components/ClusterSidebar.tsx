import { useCluster, useClusterLoading } from '@/hooks/useCluster'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { fetchCluster } from '@/services/cluster'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'

export default function ClusterSidebar({ style }: IComposedSidebarProps) {
    const { clusterName } = useParams<{ clusterName: string }>()
    const clusterInfo = useQuery(`fetchCluster:${clusterName}`, () => fetchCluster(clusterName))
    const { cluster, setCluster } = useCluster()
    const { organization, setOrganization } = useOrganization()
    const { setClusterLoading } = useClusterLoading()
    useEffect(() => {
        setClusterLoading(clusterInfo.isLoading)
        if (clusterInfo.isSuccess) {
            if (clusterInfo.data.uid !== cluster?.uid) {
                setCluster(clusterInfo.data)
            }
            if (clusterInfo.data.organization?.uid !== organization?.uid) {
                setOrganization(clusterInfo.data.organization)
            }
        } else if (clusterInfo.isLoading) {
            setCluster(undefined)
        }
    }, [
        cluster?.uid,
        clusterInfo.data,
        clusterInfo.isLoading,
        clusterInfo.isSuccess,
        organization?.uid,
        setCluster,
        setClusterLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/clusters/${clusterName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('yatai component')]),
                path: `/clusters/${clusterName}/yatai_components`,
                icon: resourceIconMapping.yatai_component,
            },
            {
                title: t('sth list', [t('deployment')]),
                path: `/clusters/${clusterName}/deployments`,
                icon: resourceIconMapping.deployment,
            },
            {
                title: t('sth list', [t('member')]),
                path: `/clusters/${clusterName}/members`,
                icon: resourceIconMapping.user_group,
            },
        ],
        [clusterName, t]
    )
    return (
        <BaseSidebar
            title={clusterName}
            icon={resourceIconMapping.cluster}
            navItems={navItems}
            style={style}
            settingsPath={`/clusters/${clusterName}/settings`}
        />
    )
}
