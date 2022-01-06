import useTranslation from '@/hooks/useTranslation'
import { createOrganizationMembers } from '@/services/organization_member'
import { useCallback, useState } from 'react'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import Card from '@/components/Card'
import MemberForm from '@/components/MemberForm'
import { ICreateMembersSchema } from '@/schemas/member'
import User from '@/components/User'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'

export default function OrganizationMemberListCard() {
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createOrganizationMembers(data)
            await membersInfo.refetch()
            setIsCreateMembersOpen(false)
        },
        [membersInfo]
    )

    return (
        <Card
            title={t('members')}
            titleIcon={resourceIconMapping.user_group}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateMembersOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={membersInfo.isLoading}
                columns={[t('user'), t('role')]}
                data={
                    membersInfo.data?.map((member) => [<User key={member.uid} user={member.user} />, t(member.role)]) ??
                    []
                }
            />
            <Modal
                isOpen={isCreateMembersOpen}
                onClose={() => setIsCreateMembersOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('member')])}</ModalHeader>
                <ModalBody>
                    <MemberForm onSubmit={handleCreateMember} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
