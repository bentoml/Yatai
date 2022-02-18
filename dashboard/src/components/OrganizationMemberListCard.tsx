import useTranslation from '@/hooks/useTranslation'
import { createOrganizationMembers } from '@/services/organization_member'
import { createUser } from '@/services/user'
import { useCallback, useState } from 'react'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { useStyletron } from 'baseui'
import Card from '@/components/Card'
import MemberForm from '@/components/MemberForm'
import { ICreateMembersSchema } from '@/schemas/member'
import { ICreateUserSchema } from '@/schemas/user'
import User from '@/components/User'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import UserForm from './UserForm'

export default function OrganizationMemberListCard() {
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()
    const [, theme] = useStyletron()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const [isCreateUserOpen, setIsCreateUserOpen] = useState(false)
    const [isSuccessfulCreateUserOpen, setIsSuccessfulCreateUserOpen] = useState(false)
    const [isEditUserRoleOpen, setIsEditUserRoleOpen] = useState(false)
    const [isDeactivateUserOpen, setIsDeactivateUserOpen] = useState(false)

    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createOrganizationMembers(data)
            await membersInfo.refetch()
            setIsCreateMembersOpen(false)
        },
        [membersInfo]
    )
    const handleCreateUser = useCallback(
        async (data: ICreateUserSchema) => {
            await createUser(data)
            await membersInfo.refetch()
            setIsCreateUserOpen(false)
            setIsSuccessfulCreateUserOpen(true)
        },
        [membersInfo]
    )

    return (
        <Card
            title={t('members')}
            titleIcon={resourceIconMapping.user_group}
            extra={
                <div style={{ display: 'flex', gap: 8, flexDirection: 'row' }}>
                    <Button size={ButtonSize.compact} onClick={() => setIsCreateUserOpen(true)}>
                        {t('create new user')}
                    </Button>
                    <Button size={ButtonSize.compact} onClick={() => setIsCreateMembersOpen(true)}>
                        {t('assign user roles')}
                    </Button>
                </div>
            }
        >
            <Table
                isLoading={membersInfo.isLoading}
                columns={[t('user'), t('role'), t('status'), t('operation')]}
                data={
                    membersInfo.data?.map((member) => [
                        <User key={member.uid} user={member.user} />,
                        t(member.role),
                        t('active'),
                        <div
                            key={member.uid}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 8,
                            }}
                        >
                            <Button size='mini'>{t('edit user role')}</Button>
                            <Button
                                size='mini'
                                overrides={{
                                    BaseButton: {
                                        style: {
                                            ':hover': {
                                                backgroundColor: theme.colors.negative,
                                            },
                                        },
                                    },
                                }}
                            >
                                {t('deactivate')}
                            </Button>
                        </div>,
                    ]) ?? []
                }
            />
            <Modal
                isOpen={isCreateMembersOpen}
                onClose={() => setIsCreateMembersOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('assign user to role')}</ModalHeader>
                <ModalBody>
                    <MemberForm onSubmit={handleCreateMember} />
                </ModalBody>
            </Modal>
            <Modal isOpen={isCreateUserOpen} onClose={() => setIsCreateUserOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create new user')}</ModalHeader>
                <ModalBody>
                    <UserForm onSubmit={handleCreateUser} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={isSuccessfulCreateUserOpen}
                onClose={() => setIsSuccessfulCreateUserOpen(false)}
                closeable
                autoFocus
                animate
            >
                <ModalHeader>Successsss</ModalHeader>
                <ModalBody>we got it</ModalBody>
            </Modal>
            <Modal isOpen={isEditUserRoleOpen} onClose={() => setIsEditUserRoleOpen(false)} closeable animate autoFocus>
                <ModalHeader>edit user role</ModalHeader>
                <ModalBody>edit user role</ModalBody>
            </Modal>
            <Modal
                isOpen={isDeactivateUserOpen}
                onClose={() => setIsDeactivateUserOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>Deactivate user</ModalHeader>
                <ModalBody>Deactivate user</ModalBody>
            </Modal>
        </Card>
    )
}
