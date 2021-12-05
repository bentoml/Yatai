import Card from '@/components/Card'
import Table from '@/components/Table'
import User from '@/components/User'
import { useModelRepository, useModelRepositoryLoading } from '@/hooks/useModelRepository'
import useTranslation from '@/hooks/useTranslation'
import { formatDateTime } from '@/utils/datetime'
import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'

export default function ModelRepositoryOverview() {
    const { modelRepository } = useModelRepository()
    const { modelRepositoryLoading } = useModelRepositoryLoading()
    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={modelRepositoryLoading}
                columns={[t('name'), t('creator'), t('created_at')]}
                data={[
                    [
                        modelRepository?.name,
                        modelRepository?.creator && <User user={modelRepository?.creator} />,
                        modelRepository && formatDateTime(modelRepository.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
