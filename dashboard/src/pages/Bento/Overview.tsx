import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useBento, useBentoLoading } from '@/hooks/useBento'
import Card from '@/components/Card'
import { formatDateTime } from '@/utils/datetime'
import User from '@/components/User'

export default function BentoOverview() {
    const { bento } = useBento()
    const { bentoLoading } = useBentoLoading()

    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={bentoLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={[
                    [
                        bento?.name,
                        bento?.description,
                        bento?.creator && <User user={bento?.creator} />,
                        bento && formatDateTime(bento.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
