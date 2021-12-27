import React from 'react'
import { usePage } from '@/hooks/usePage'
import useTranslation from '@/hooks/useTranslation'
import { listBentoRepositoryDeployments } from '@/services/bento_repository'
import { useQuery } from 'react-query'
import { Link, useParams } from 'react-router-dom'
import qs from 'qs'
import Card from '@/components/Card'
import Table from '@/components/Table'
import DeploymentStatusTag from '@/components/DeploymentStatusTag'
import User from '@/components/User'
import Time from '@/components/Time'
import { resourceIconMapping } from '@/consts'

export default function BentoRepositoryDeployments() {
    const { bentoRepositoryName } = useParams<{ bentoRepositoryName: string }>()
    const [t] = useTranslation()
    const [page] = usePage()
    const deploymentsInfo = useQuery(`bentoRepository:${bentoRepositoryName}:deployments:${qs.stringify(page)}`, () =>
        listBentoRepositoryDeployments(bentoRepositoryName, page)
    )

    return (
        <Card title={t('deployments')} titleIcon={resourceIconMapping.deployment}>
            <Table
                isLoading={deploymentsInfo.isLoading}
                columns={[
                    t('name'),
                    t('cluster'),
                    t('status'),
                    t('creator'),
                    t('Time since Creation'),
                    t('updated_at'),
                ]}
                data={
                    deploymentsInfo.data?.items.map((deployment) => {
                        return [
                            <Link
                                key={deployment.uid}
                                to={`/clusters/${deployment.cluster?.name}/deployments/${deployment.name}`}
                            >
                                {deployment.name}
                            </Link>,
                            <Link key={deployment.cluster?.uid} to={`/clusters/${deployment.cluster?.name}`}>
                                {deployment.cluster?.name}
                            </Link>,
                            <DeploymentStatusTag key={deployment.uid} status={deployment.status} />,
                            deployment?.creator && <User user={deployment.creator} />,
                            deployment?.created_at && <Time time={deployment.created_at} />,
                            deployment?.latest_revision && <Time time={deployment.latest_revision.created_at} />,
                        ]
                    }) ?? []
                }
                paginationProps={{
                    start: deploymentsInfo.data?.start,
                    count: deploymentsInfo.data?.count,
                    total: deploymentsInfo.data?.total,
                    afterPageChange: () => {
                        deploymentsInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
