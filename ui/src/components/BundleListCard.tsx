import React, { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { GoPackage } from 'react-icons/go'
import { createBundle, listBundles } from '@/services/bundle'
import { usePage } from '@/hooks/usePage'
import { ICreateBundleSchema } from '@/schemas/bundle'
import BundleForm from '@/components/BundleForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'

export interface IBundleListCardProps {
    orgName: string
    clusterName: string
}

export default function BundleListCard({ orgName, clusterName }: IBundleListCardProps) {
    const [page, setPage] = usePage()
    const bundlesInfo = useQuery(`fetchClusterBundles:${orgName}:${clusterName}`, () =>
        listBundles(orgName, clusterName, page)
    )
    const [isCreateBundleOpen, setIsCreateBundleOpen] = useState(false)
    const handleCreateBundle = useCallback(
        async (data: ICreateBundleSchema) => {
            await createBundle(orgName, clusterName, data)
            await bundlesInfo.refetch()
            setIsCreateBundleOpen(false)
        },
        [bundlesInfo, clusterName, orgName]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('bundle')])}
            titleIcon={GoPackage}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBundleOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bundlesInfo.isLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={
                    bundlesInfo.data?.items.map((bundle) => [
                        <Link key={bundle.uid} to={`/orgs/${orgName}/clusters/${clusterName}/bundles/${bundle.name}`}>
                            {bundle.name}
                        </Link>,
                        bundle.description,
                        bundle.creator && <User user={bundle.creator} />,
                        formatTime(bundle.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bundlesInfo.data?.start,
                    count: bundlesInfo.data?.count,
                    total: bundlesInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bundlesInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateBundleOpen} onClose={() => setIsCreateBundleOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('bundle')])}</ModalHeader>
                <ModalBody>
                    <BundleForm onSubmit={handleCreateBundle} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
