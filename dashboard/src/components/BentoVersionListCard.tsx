import React, { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { createBentoVersion, listBentoVersions } from '@/services/bento_version'
import { usePage } from '@/hooks/usePage'
import { ICreateBentoVersionSchema } from '@/schemas/bento_version'
import BentoVersionForm from '@/components/BentoVersionForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'

export interface IBentoVersionListCardProps {
    orgName: string
    bentoName: string
}

export default function BentoVersionListCard({ orgName, bentoName }: IBentoVersionListCardProps) {
    const [page, setPage] = usePage()
    const bentoVersionsInfo = useQuery(`fetchClusterBentoVersions:${orgName}:${bentoName}`, () =>
        listBentoVersions(orgName, bentoName, page)
    )
    const [isCreateBentoVersionOpen, setIsCreateBentoVersionOpen] = useState(false)
    const handleCreateBentoVersion = useCallback(
        async (data: ICreateBentoVersionSchema) => {
            await createBentoVersion(orgName, bentoName, data)
            await bentoVersionsInfo.refetch()
            setIsCreateBentoVersionOpen(false)
        },
        [bentoName, bentoVersionsInfo, orgName]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('version')])}
            titleIcon={resourceIconMapping.bento_version}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoVersionOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bentoVersionsInfo.isLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={
                    bentoVersionsInfo.data?.items.map((bentoVersion) => [
                        <Link
                            key={bentoVersion.uid}
                            to={`/orgs/${orgName}/bentos/${bentoName}/versions/${bentoVersion.version}`}
                        >
                            {bentoVersion.version}
                        </Link>,
                        bentoVersion.description,
                        bentoVersion.creator && <User user={bentoVersion.creator} />,
                        formatTime(bentoVersion.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentoVersionsInfo.data?.start,
                    count: bentoVersionsInfo.data?.count,
                    total: bentoVersionsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bentoVersionsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateBentoVersionOpen}
                onClose={() => setIsCreateBentoVersionOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('version')])}</ModalHeader>
                <ModalBody>
                    <BentoVersionForm onSubmit={handleCreateBentoVersion} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
