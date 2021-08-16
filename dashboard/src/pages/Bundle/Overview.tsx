import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useBundle, useBundleLoading } from '@/hooks/useBundle'
import Card from '@/components/Card'
import { formatTime } from '@/utils/datetime'
import User from '@/components/User'

export default function BundleOverview() {
    const { bundle } = useBundle()
    const { bundleLoading } = useBundleLoading()

    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={bundleLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={[
                    [
                        bundle?.name,
                        bundle?.description,
                        bundle?.creator && <User user={bundle?.creator} />,
                        bundle && formatTime(bundle.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
