import { useCluster, useClusterLoading } from '@/hooks/useCluster'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useQuery } from 'react-query'
import { useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { fetchCluster } from '@/services/cluster'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import { AiOutlineSetting } from 'react-icons/ai'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import { useFetchInfo } from '@/hooks/useFetchInfo'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import BaseSubLayout from './BaseSubLayout'

export interface IClusterLayoutProps {
    children: React.ReactNode
}

export default function ClusterLayout({ children }: IClusterLayoutProps) {
    const { clusterName } = useParams<{ clusterName: string }>()
    const clusterInfo = useQuery(`fetchCluster:${clusterName}`, () => fetchCluster(clusterName))
    const { cluster, setCluster } = useCluster()
    const { organization, setOrganization } = useOrganization()
    const { setClusterLoading } = useClusterLoading()
    const membersInfo = useFetchOrganizationMembers()
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

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('clusters'),
                path: '/clusters',
                icon: resourceIconMapping.cluster,
            },
            {
                title: clusterName,
                path: `/clusters/${clusterName}`,
            },
        ],
        [clusterName, t]
    )

    const { currentUser } = useCurrentUser()
    const infoInfo = useFetchInfo()
    const showHiddenNavItems = useMemo(() => {
        return (
            infoInfo.data?.is_sass &&
            membersInfo.data?.find((m) => {
                return m.user.uid === currentUser?.uid
            })?.role === 'admin'
        )
    }, [currentUser?.uid, infoInfo.data?.is_sass, membersInfo.data])

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/clusters/${clusterName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('yatai components'),
                path: `/clusters/${clusterName}/yatai_components`,
                icon: resourceIconMapping.yatai_component,
            },
            {
                title: t('deployments'),
                path: `/clusters/${clusterName}/deployments`,
                icon: resourceIconMapping.deployment,
            },
            {
                title: t('members'),
                path: `/clusters/${clusterName}/members`,
                icon: resourceIconMapping.user_group,
                hidden: !showHiddenNavItems,
            },
            {
                title: t('settings'),
                path: `/clusters/${clusterName}/settings`,
                icon: AiOutlineSetting,
            },
        ],
        [clusterName, showHiddenNavItems, t]
    )
    return (
        <BaseSubLayout breadcrumbItems={breadcrumbItems} navItems={navItems}>
            {children}
        </BaseSubLayout>
    )
}
