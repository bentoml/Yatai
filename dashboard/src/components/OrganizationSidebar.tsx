import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'
import React, { useMemo } from 'react'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'
import { FiActivity } from 'react-icons/fi'
import { useFetchOrganizationYataiComponents } from '@/hooks/useFetchYataiComponents'

export default function OrganizationSidebar({ style }: IComposedSidebarProps) {
    const { yataiComponentsInfo } = useFetchOrganizationYataiComponents()

    const deploymentDisabled = useMemo(() => {
        return yataiComponentsInfo.data?.find((c) => c.name === 'deployment') === undefined
    }, [yataiComponentsInfo.data])

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
                disabled: deploymentDisabled,
                helpMessage: deploymentDisabled
                    ? t('you need to install yatai-deployment component to enable deployment function')
                    : undefined,
                activePathPattern:
                    /^\/(deployments|new_deployment|clusters\/[^/]+\/namespaces\/[^/]+\/deployments\/[^/]+)\/?/,
            },
            {
                title: t('clusters'),
                path: '/clusters',
                icon: resourceIconMapping.cluster,
            },
            {
                title: t('events'),
                path: '/events',
                icon: FiActivity,
            },
        ],
        [deploymentDisabled, t]
    )
    return <BaseSidebar navItems={navItems} style={style} />
}
