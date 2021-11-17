import React, { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { createOrganization } from '@/services/organization'
import { usePage } from '@/hooks/usePage'
import { ICreateOrganizationSchema } from '@/schemas/organization'
import OrganizationForm from '@/components/OrganizationForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { useFetchOrganizations } from '@/hooks/useFetchOrganizations'
import { resourceIconMapping } from '@/consts'

export default function OrganizationListCard() {
    const [page] = usePage()
    const organizationsInfo = useFetchOrganizations(page)
    const [isCreateOrganizationOpen, setIsCreateOrganizationOpen] = useState(false)
    const handleCreateOrganization = useCallback(
        async (data: ICreateOrganizationSchema) => {
            await createOrganization(data)
            await organizationsInfo.refetch()
            setIsCreateOrganizationOpen(false)
        },
        [organizationsInfo]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('organization')])}
            titleIcon={resourceIconMapping.organization}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateOrganizationOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={organizationsInfo.isLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={
                    organizationsInfo.data?.items.map((organization) => {
                        return [
                            // TODO
                            <Link key={organization.uid} to='i dont known'>
                                {organization.name}
                            </Link>,
                            organization.description,
                            organization.creator && <User user={organization.creator} />,
                            formatDateTime(organization.created_at),
                        ]
                    }) ?? []
                }
                paginationProps={{
                    start: organizationsInfo.data?.start,
                    count: organizationsInfo.data?.count,
                    total: organizationsInfo.data?.total,
                    afterPageChange: () => {
                        organizationsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateOrganizationOpen}
                onClose={() => setIsCreateOrganizationOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('organization')])}</ModalHeader>
                <ModalBody>
                    <OrganizationForm onSubmit={handleCreateOrganization} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
