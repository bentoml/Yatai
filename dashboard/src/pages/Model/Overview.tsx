import Card from '@/components/Card'
import Table from '@/components/Table'
import User from '@/components/User'
import { useModel, useModelLoading } from '@/hooks/useModel'
import useTranslation from '@/hooks/useTranslation'
import { formatDateTime } from '@/utils/datetime'
import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'

export default function ModelOverview() {
    const { model } = useModel()
    const { modelLoading } = useModelLoading()
    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={modelLoading}
                columns={[t('name'), t('creator'), t('created_at')]}
                data={[
                    [
                        model?.name,
                        model?.creator && <User user={model?.creator} />,
                        model && formatDateTime(model.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
