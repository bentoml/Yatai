import useTranslation from '@/hooks/useTranslation'
import { createClusterMembers, listClusterMembers } from '@/services/cluster_member'
import React, { useCallback, useState } from 'react'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import MemberForm from '@/components/MemberForm'
import { ICreateMembersSchema } from '@/schemas/member'
import User from '@/components/User'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'

export interface IClusterMemberListCardProps {
    orgName: string
    clusterName: string
}

export default function ClusterMemberListCard({ orgName, clusterName }: IClusterMemberListCardProps) {
    const membersInfo = useQuery(`fetchClusterMembers:${orgName}:${clusterName}`, () =>
        listClusterMembers(orgName, clusterName)
    )
    const [t] = useTranslation()
    const [isCreateMembersOpen, setIsCreateMembersOpen] = useState(false)
    const handleCreateMember = useCallback(
        async (data: ICreateMembersSchema) => {
            await createClusterMembers(orgName, clusterName, data)
            await membersInfo.refetch()
            setIsCreateMembersOpen(false)
        },
        [clusterName, membersInfo, orgName]
    )

    return (
        <>
            <Card
                title={t('sth list', [t('member')])}
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
