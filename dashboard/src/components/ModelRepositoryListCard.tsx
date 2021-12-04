import { usePage } from '@/hooks/usePage'
import useTranslation from '@/hooks/useTranslation'
import { ICreateModelRepositorySchema } from '@/schemas/model_repository'
import { createModelRepository, listModelRepositories } from '@/services/model_repository'
import React, { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { resourceIconMapping } from '@/consts'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import Table from '@/components/Table'
import User from '@/components/User'
import { formatDateTime } from '@/utils/datetime'
import { Link } from 'react-router-dom'
import qs from 'qs'

export default function ModelRepositoryListCard() {
    const [page] = usePage()
    const modelRepositoriesInfo = useQuery(`fetchModelRepositories:${qs.stringify(page)}`, () =>
        listModelRepositories(page)
    )
    const [isCreateModelOpen, setIsCreateModelOpen] = useState(false)
    // eslint-disable-next-line
    const handleCreateModel = useCallback(
        async (data: ICreateModelRepositorySchema) => {
            await createModelRepository(data)
            await modelRepositoriesInfo.refetch()
            setIsCreateModelOpen(false)
        },
        [modelRepositoriesInfo]
    )
    const [t] = useTranslation()
    return (
        <Card
            title={t('sth list', [t('model repository')])}
            titleIcon={resourceIconMapping.model}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateModelOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={modelRepositoriesInfo.isLoading}
                columns={[t('name'), t('version'), t('created_at')]}
                data={
                    modelRepositoriesInfo.data?.items?.map((modelRepository) => [
                        <Link key={modelRepository.uid} to={`/model_repositories/${modelRepository.name}`}>
                            {modelRepository.name}
                        </Link>,
                        modelRepository.latest_model?.version,
                        modelRepository.creator && <User user={modelRepository.creator} />,
                        formatDateTime(modelRepository.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: modelRepositoriesInfo.data?.start,
                    count: modelRepositoriesInfo.data?.count,
                    total: modelRepositoriesInfo.data?.total,
                    afterPageChange: () => {
                        modelRepositoriesInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateModelOpen} onClose={() => setIsCreateModelOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('model repository')])}</ModalHeader>
                <ModalBody>Model form creat model</ModalBody>
            </Modal>
        </Card>
    )
}
