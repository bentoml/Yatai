import { useFetchModel } from '@/hooks/useFetchModel'
import { useModel, useModelLoading } from '@/hooks/useModel'
import useTranslation from '@/hooks/useTranslation'
import React, { useEffect, useMemo } from 'react'
import { useParams } from 'react-router-dom'
import { INavItem } from '@/components/BaseSidebar'
import { resourceIconMapping } from '@/consts'
import BaseSubLayout from './BaseSubLayout'

export interface IModelLayoutProps {
    children: React.ReactNode
}

export default function ModelLayout({ children }: IModelLayoutProps) {
    const { modelRepositoryName, modelVersion } = useParams<{ modelRepositoryName: string; modelVersion: string }>()
    const modelInfo = useFetchModel(modelRepositoryName, modelVersion)
    const { setModel } = useModel()
    const { setModelLoading } = useModelLoading()
    useEffect(() => {
        setModel(modelInfo.data)
        setModelLoading(modelInfo.isLoading)
    }, [modelInfo, setModel, setModelLoading])

    const [t] = useTranslation()

    const breadcrumbItems: INavItem[] = useMemo(
        () => [
            {
                title: t('model repositories'),
                path: '/model_repositories',
                icon: resourceIconMapping.bento,
            },
            {
                title: modelRepositoryName,
                path: `/model_repositories/${modelRepositoryName}`,
            },
            {
                title: modelVersion,
                path: `/model_repositories/${modelRepositoryName}/models/${modelVersion}`,
            },
        ],
        [modelRepositoryName, modelVersion, t]
    )

    return <BaseSubLayout breadcrumbItems={breadcrumbItems}>{children}</BaseSubLayout>
}
