import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useCluster, useClusterLoading } from '@/hooks/useCluster'
import Card from '@/components/Card'
import { formatTime } from '@/utils/datetime'
import User from '@/components/User'

export default function ClusterOverview() {
    const { cluster } = useCluster()
    const { clusterLoading } = useClusterLoading()

    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={clusterLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={[
                    [
                        cluster?.name,
                        cluster?.description,
                        cluster?.creator && <User user={cluster?.creator} />,
                        cluster && formatTime(cluster.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
