import { useQuery } from 'react-query'
import BaseSidebar, { IComposedSidebarProps, INavItem } from '@/components/BaseSidebar'
import { useParams } from 'react-router-dom'
import React, { useEffect, useMemo } from 'react'
import { fetchModelRepository } from '@/services/model_repository'
import { useModelRepository, useModelRepositoryLoading } from '@/hooks/useModelRepository'
import { useOrganization } from '@/hooks/useOrganization'
import { resourceIconMapping } from '@/consts'
import useTranslation from '@/hooks/useTranslation'
import { RiSurveyLine } from 'react-icons/ri'

export default function ModelRepositorySidebar({ style }: IComposedSidebarProps) {
    // eslint-disable-line
    const { modelRepositoryName } = useParams<{ modelRepositoryName: string }>()
    const modelInfo = useQuery(`fetchModelRepository:${modelRepositoryName}`, () =>
        fetchModelRepository(modelRepositoryName)
    )
    const { modelRepository: model, setModelRepository: setModel } = useModelRepository()
    const { organization, setOrganization } = useOrganization()
    const { setModelRepositoryLoading: setModelLoading } = useModelRepositoryLoading()
    useEffect(() => {
        setModelLoading(modelInfo.isLoading)
        if (modelInfo.isSuccess) {
            if (modelInfo.data.uid !== model?.uid) {
                setModel(modelInfo.data)
            }
            if (modelInfo.data.organization?.uid !== organization?.uid) {
                setOrganization(modelInfo.data.organization)
            }
        } else if (modelInfo.isLoading) {
            setModel(undefined)
        }
    }, [
        model?.uid,
        modelInfo.data,
        modelInfo.isLoading,
        modelInfo.isSuccess,
        organization?.uid,
        setModel,
        setModelLoading,
        setOrganization,
    ])
    const [t] = useTranslation()

    const navItems: INavItem[] = useMemo(
        () => [
            {
                title: t('overview'),
                path: `/model_repositories/${modelRepositoryName}`,
                icon: RiSurveyLine,
            },
            {
                title: t('sth list', [t('model')]),
                path: `/model_repositories/${modelRepositoryName}/models`,
                icon: resourceIconMapping.model,
            },
        ],
        [modelRepositoryName, t]
    )

    return (
        <BaseSidebar style={style} title={modelRepositoryName} icon={resourceIconMapping.bento} navItems={navItems} />
    )
}
