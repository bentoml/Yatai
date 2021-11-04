import React, { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { createBento, listBentos } from '@/services/bento'
import { usePage } from '@/hooks/usePage'
import { ICreateBentoSchema } from '@/schemas/bento'
import BentoForm from '@/components/BentoForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'

export default function BentoListCard() {
    const [page, setPage] = usePage()
    const bentosInfo = useQuery('fetchClusterBentos', () => listBentos(page))
    const [isCreateBentoOpen, setIsCreateBentoOpen] = useState(false)
    const handleCreateBento = useCallback(
        async (data: ICreateBentoSchema) => {
            await createBento(data)
            await bentosInfo.refetch()
            setIsCreateBentoOpen(false)
        },
        [bentosInfo]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('bento')])}
            titleIcon={resourceIconMapping.bento}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={bentosInfo.isLoading}
                columns={[t('name'), t('latest version'), t('creator'), t('created_at')]}
                data={
                    bentosInfo.data?.items.map((bento) => [
                        <Link key={bento.uid} to={`/bentos/${bento.name}`}>
                            {bento.name}
                        </Link>,
                        bento.latest_version?.version,
                        bento.creator && <User user={bento.creator} />,
                        formatTime(bento.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentosInfo.data?.start,
                    count: bentosInfo.data?.count,
                    total: bentosInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bentosInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateBentoOpen} onClose={() => setIsCreateBentoOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('bento')])}</ModalHeader>
                <ModalBody>
                    <BentoForm onSubmit={handleCreateBento} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
