import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useBentoRepository, useBentoRepositoryLoading } from '@/hooks/useBentoRepository'
import Card from '@/components/Card'
import { formatDateTime } from '@/utils/datetime'
import User from '@/components/User'

export default function BentoRepositoryOverview() {
    const { bentoRepository } = useBentoRepository()
    const { bentoRepositoryLoading } = useBentoRepositoryLoading()

    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={bentoRepositoryLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={[
                    [
                        bentoRepository?.name,
                        bentoRepository?.description,
                        bentoRepository?.creator && <User user={bentoRepository?.creator} />,
                        bentoRepository && formatDateTime(bentoRepository.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
