import useTranslation from '@/hooks/useTranslation'
import { createOrganizationMembers, listOrganizationMembers } from '@/services/organization_member'
import React, { useCallback, useState } from 'react'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { HiOutlineUserGroup } from 'react-icons/hi'
import MemberForm from '@/components/MemberForm'
import { ICreateMembersSchema } from '@/schemas/member'
import User from '@/components/User'
import Table from '@/components/Table'

export interface IOrganizationMemberListCardProps {
    orgName: string
}

export default function OrganizationMemberListCard({ orgName }: IOrganizationMemberListCardProps) {
    const membersInfo = useQuery(`fetchOrgMembers:${orgName}`, () => listOrganizationMembers(orgName))
    const [t] = useTranslation()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createOrganizationMembers(orgName, data)
            await membersInfo.refetch()
            setIsCreateMembersOpen(false)
        },
        [membersInfo, orgName]
    )

    return (
        <>
            <Card
                title={t('sth list', [t('member')])}
                titleIcon={HiOutlineUserGroup}
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
                        membersInfo.data?.map((member) => [
                            <User key={member.uid} user={member.user} />,
                            t(member.role),
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
                    <ModalHeader>{t('create sth', [t('member')])}</ModalHeader>
                    <ModalBody>
                        <MemberForm onSubmit={handleCreateMember} />
                    </ModalBody>
                </Modal>
            </Card>
        </>
    )
}
