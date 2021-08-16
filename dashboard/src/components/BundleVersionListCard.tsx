import React, { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { createBundleVersion, listBundleVersions } from '@/services/bundle_version'
import { usePage } from '@/hooks/usePage'
import { ICreateBundleVersionSchema } from '@/schemas/bundle_version'
import BundleVersionForm from '@/components/BundleVersionForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'

export interface IBundleVersionListCardProps {
    orgName: string
    bundleName: string
}

export default function BundleVersionListCard({ orgName, bundleName }: IBundleVersionListCardProps) {
    const [page, setPage] = usePage()
    const bundleVersionsInfo = useQuery(`fetchClusterBundleVersions:${orgName}:${bundleName}`, () =>
        listBundleVersions(orgName, bundleName, page)
    )
    const [isCreateBundleVersionOpen, setIsCreateBundleVersionOpen] = useState(false)
    const handleCreateBundleVersion = useCallback(
        async (data: ICreateBundleVersionSchema) => {
            await createBundleVersion(orgName, bundleName, data)
            await bundleVersionsInfo.refetch()
            setIsCreateBundleVersionOpen(false)
        },
        [bundleName, bundleVersionsInfo, orgName]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.bundle_version}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBundleVersionOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bundleVersionsInfo.isLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={
                    bundleVersionsInfo.data?.items.map((bundleVersion) => [
                        <Link
                            key={bundleVersion.uid}
                            to={`/orgs/${orgName}/bundles/${bundleName}/versions/${bundleVersion.version}`}
                        >
                            {bundleVersion.version}
                        </Link>,
                        bundleVersion.description,
                        bundleVersion.creator && <User user={bundleVersion.creator} />,
                        formatTime(bundleVersion.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bundleVersionsInfo.data?.start,
                    count: bundleVersionsInfo.data?.count,
                    total: bundleVersionsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bundleVersionsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateBundleVersionOpen}
                onClose={() => setIsCreateBundleVersionOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('version')])}</ModalHeader>
                <ModalBody>
                    <BundleVersionForm onSubmit={handleCreateBundleVersion} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
