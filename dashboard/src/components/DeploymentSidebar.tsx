import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import { useCluster } from '@/hooks/useCluster'
import { useFetchDeployment } from '@/hooks/useFetchDeployment'

export default function DeploymentSidebar({ style }: IComposedSidebarProps) {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const { deploymentInfo } = useFetchDeployment(clusterName, kubeNamespace, deploymentName)
    const { deployment, setDeployment } = useDeployment()
    const { organization, setOrganization } = useOrganization()
    const { cluster, setCluster } = useCluster()
    const { setDeploymentLoading } = useDeploymentLoading()

    useEffect(() => {
        setDeploymentLoading(deploymentInfo.isLoading)
        if (deploymentInfo.isSuccess) {
            if (deploymentInfo.data.uid !== deployment?.uid) {
                setDeployment(deploymentInfo.data)
            }
            if (deploymentInfo.data.cluster?.uid !== cluster?.uid) {
                setCluster(deploymentInfo.data.cluster)
            }
            if (deploymentInfo.data.cluster?.organization?.uid !== organization?.uid) {
                setOrganization(deploymentInfo.data.cluster?.organization)
            }
        } else if (deploymentInfo.isLoading) {
            setDeployment(undefined)
        }
    }, [
        cluster?.uid,
        deployment?.uid,
        deploymentInfo.data,
        deploymentInfo.isLoading,
        deploymentInfo.isSuccess,
        organization?.uid,
        setCluster,
        setDeployment,
        setDeploymentLoading,
        setOrganization,
    ])

    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('revisions'),
                path: `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/revisions`,
                icon: resourceIconMapping.deployment_revision,
            },
        ],
        [clusterName, deploymentName, kubeNamespace, t]
    )
    return (
        <BaseSidebar title={deploymentName} icon={resourceIconMapping.deployment} navItems={navItems} style={style} />
    )
}
