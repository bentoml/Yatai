import useTranslation from '@/hooks/useTranslation'
import _ from 'lodash'
import { createOrganizationMembers } from '@/services/organization_member'
import { registerUser } from '@/services/user' // eslint-disable-line
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
import RegisterUserForm from './RegisterUserForm'

export default function OrganizationMemberListCard() {
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()
    const [, theme] = useStyletron()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const [isRegisterMemberOpen, setIsRegisterMemberOpen] = useState(false)
    const [isSuccessfulRegisterOpen, setIsSuccessfulRegisterOpen] = useState(false)

    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createOrganizationMembers(data)
            // await membersInfo.refetch()
            setIsCreateMembersOpen(false)
        },
        [membersInfo]
    )
    const handleRegisterUser = useCallback(
        async (data: ICreateUserSchema) => {
            /* Currently, it is a two step process:
             * 1. Register a new user
             * 2. Add the user to the organization
             */
            const registerData = _.omit(data, ['role'])
            const membershipData = { role: data.role, usernames: [data.name] }
            await registerUser(registerData)
            debugger // eslint-disable-line
            console.log('created user') // eslint-disable-line
            await createOrganizationMembers(membershipData)
            setIsRegisterMemberOpen(false)
            setIsSuccessfulRegisterOpen(true)
            await membersInfo.refetch()
        },
        [membersInfo]
    )

    return (
        <Card
            title={t('members')}
            titleIcon={resourceIconMapping.user_group}
            extra={
                <div style={{ display: 'flex', gap: 8, flexDirection: 'row' }}>
                    <Button size={ButtonSize.compact} onClick={() => setIsRegisterMemberOpen(true)}>
                        {t('register new user')}
                    </Button>
                    <Button size={ButtonSize.compact} onClick={() => setIsCreateMembersOpen(true)}>
                        {t('assign user')}
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
            <Modal
                isOpen={isRegisterMemberOpen}
                onClose={() => setIsRegisterMemberOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('register new user')}</ModalHeader>
                <ModalBody>
                    <RegisterUserForm onSubmit={handleRegisterUser} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={isSuccessfulRegisterOpen}
                onClose={() => setIsSuccessfulRegisterOpen(false)}
                closeable
                autoFocus
                animate
            >
                <ModalHeader>Successsss</ModalHeader>
                <ModalBody>we got it</ModalBody>
            </Modal>
        </Card>
    )
}
