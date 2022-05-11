import useTranslation from '@/hooks/useTranslation'
import { createOrganizationMembers } from '@/services/organization_member'
import { createUser } from '@/services/user'
import { useCallback, useState } from 'react'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { toaster } from 'baseui/toast'
import { generate } from 'generate-password'
import CopyToClipboard from 'react-copy-to-clipboard'
import { TiClipboard } from 'react-icons/ti'
import Card from '@/components/Card'
import MemberForm from '@/components/MemberForm'
import { ICreateMembersSchema } from '@/schemas/member'
import { ICreateUserSchema } from '@/schemas/user'
import User from '@/components/User'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { IOrganizationMemberSchema } from '@/schemas/organization_member'
import { useStyletron } from 'baseui'
import UserForm from './UserForm'

const isDeactivated = (deleted_at: string | undefined): boolean => {
    return !!deleted_at
}

export default function OrganizationMemberListCard() {
    const membersInfo = useFetchOrganizationMembers()
    const [t] = useTranslation()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const [isCreateUserOpen, setIsCreateUserOpen] = useState(false)
    const [isSuccessfulCreateUserOpen, setIsSuccessfulCreateUserOpen] = useState(false)
    const [isEditUserRoleOpen, setIsEditUserRoleOpen] = useState(false)
    const [selectedMember, setSelectedMember] = useState<IOrganizationMemberSchema | undefined>(undefined)
    const [newUserInfo, setNewUserInfo] = useState<ICreateUserSchema | undefined>(undefined)
    const [displaySuccessCopiedMessage, setDisplaySuccessCopiedMessage] = useState(false)
    const [, theme] = useStyletron()

    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createOrganizationMembers(data)
            await membersInfo.refetch()
            setIsCreateMembersOpen(false)
            setIsEditUserRoleOpen(false)
            toaster.positive(t('assigned new role'), { autoHideDuration: 2000 })
        },
        [t, membersInfo]
    )
    const handleCreateUser = useCallback(
        async (data: ICreateUserSchema) => {
            const newData = { ...data, password: generate({ length: 10, numbers: true }) }
            await createUser(newData)
            await membersInfo.refetch()
            setIsCreateUserOpen(false)
            setIsSuccessfulCreateUserOpen(true)
            setNewUserInfo(newData)
            toaster.positive(`${t('created new user')} ${data.name}`, { autoHideDuration: 2000 })
        },
        [t, membersInfo]
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
                        isDeactivated(member.deleted_at) ? t('deactivated') : t('active'),
                        <div
                            key={member.uid}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 8,
                            }}
                        >
                            <Button
                                size='mini'
                                onClick={() => {
                                    setIsEditUserRoleOpen(true)
                                    setSelectedMember(member)
                                }}
                            >
                                {t('edit user role')}
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
                onClose={() => {
                    setIsSuccessfulCreateUserOpen(false)
                    setNewUserInfo(undefined)
                    setDisplaySuccessCopiedMessage(false)
                }}
                closeable
                autoFocus
                animate
            >
                <ModalHeader>{t('success')}</ModalHeader>
                <ModalBody>
                    <div
                        style={{
                            marginBottom: 20,
                        }}
                    >
                        <p>You can view and copy the login information below:</p>
                        <pre
                            style={{
                                lineHeight: '1.6',
                                fontSize: '12px',
                                padding: '4px 8px',
                                backgroundColor: theme.colors.backgroundTertiary,
                                borderRadius: 4,
                            }}
                        >
                            <div>Sign-in URL: {window.location.origin}/login</div>
                            <div>Email: {newUserInfo?.email}</div>
                            <div>Password: {newUserInfo?.password}</div>
                        </pre>
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                        }}
                    >
                        <div style={{ flexGrow: 1 }} />
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                            }}
                        >
                            <div style={{ display: 'flex', flexDirection: 'row' }}>
                                <div style={{ flexGrow: 1 }} />
                                <CopyToClipboard
                                    text={`Sign-in URL: ${window.location.origin}/login
Email: ${newUserInfo?.email}
Password: ${newUserInfo?.password}`}
                                    onCopy={() => {
                                        setDisplaySuccessCopiedMessage(true)
                                    }}
                                >
                                    <Button startEnhancer={<TiClipboard size={14} />} kind='secondary' size='compact'>
                                        {t('copy')}
                                    </Button>
                                </CopyToClipboard>
                            </div>
                            {displaySuccessCopiedMessage && (
                                <div style={{ marginTop: 8 }}>{t('copied to clipboard')}</div>
                            )}
                        </div>
                    </div>
                </ModalBody>
            </Modal>
            <Modal isOpen={isEditUserRoleOpen} onClose={() => setIsEditUserRoleOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('edit user role')}</ModalHeader>
                <ModalBody>
                    <MemberForm member={selectedMember} onSubmit={handleCreateMember} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
