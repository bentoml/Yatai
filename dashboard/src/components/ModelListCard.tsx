import { usePage } from '@/hooks/usePage'
import useTranslation from '@/hooks/useTranslation'
import { ICreateModelSchema } from '@/schemas/model'
import { createModel, listModels } from '@/services/model'
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

export default function ModelListCard() {
    const [page] = usePage()
    const modelInfo = useQuery(`fetchModels:${qs.stringify(page)}`, () => listModels(page))
    const [isCreateModelOpen, setIsCreateModelOpen] = useState(false)
    // eslint-disable-next-line
    const handleCreateModel = useCallback(
        async (data: ICreateModelSchema) => {
            await createModel(data)
            await modelInfo.refetch()
            setIsCreateModelOpen(false)
        },
        [modelInfo]
    )
    const [t] = useTranslation()
    return (
        <Card
            title={t('sth list', [t('model')])}
            titleIcon={resourceIconMapping.model}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateModelOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={modelInfo.isLoading}
                columns={[t('name'), t('version'), t('created_at')]}
                data={
                    modelInfo.data?.items?.map((model) => [
                        <Link key={model.uid} to={`/models/${model.name}`}>
                            {model.name}
                        </Link>,
                        model.latest_version?.version,
                        model.creator && <User user={model.creator} />,
                        formatDateTime(model.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: modelInfo.data?.start,
                    count: modelInfo.data?.count,
                    total: modelInfo.data?.total,
                    afterPageChange: () => {
                        modelInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateModelOpen} onClose={() => setIsCreateModelOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('model')])}</ModalHeader>
                <ModalBody>Model form creat model</ModalBody>
            </Modal>
        </Card>
    )
}
