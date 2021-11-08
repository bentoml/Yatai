import { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { updateDeployment } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import DeploymentForm from '@/components/DeploymentForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { IUpdateDeploymentSchema } from '@/schemas/deployment'
import { useDeployment } from '@/hooks/useDeployment'
import { Theme } from 'baseui/theme'
import color from 'color'
import { StyledLink } from 'baseui/link'
import { IDeploymentSnapshotSchema } from '@/schemas/deployment_snapshot'
import { useFetchDeploymentSnapshots } from '@/hooks/useFetchDeploymentSnapshots'
import DeploymentSnapshotDetail from './DeploymentSnapshotDetail'

export interface IDeploymentSnapshotListCardProps {
    clusterName: string
    deploymentName: string
}

export default function DeploymentSnapshotListCard({ clusterName, deploymentName }: IDeploymentSnapshotListCardProps) {
    const [page, setPage] = usePage()
    const { deployment } = useDeployment()
    const [desiredShowDeploymentSnapshot, setDesiredShowDeploymentSnapshot] = useState<IDeploymentSnapshotSchema>()
    const { deploymentSnapshotsInfo } = useFetchDeploymentSnapshots(clusterName, deploymentName, page)
    const [isCreateDeploymentSnapshotOpen, setIsCreateDeploymentSnapshotOpen] = useState(false)
    const handleCreateDeploymentSnapshot = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, deploymentName, data)
            await deploymentSnapshotsInfo.refetch()
            setIsCreateDeploymentSnapshotOpen(false)
        },
        [deploymentSnapshotsInfo, clusterName, deploymentName]
    )

    const [t] = useTranslation()

    return (
        <Card title={t('sth list', [t('snapshot')])} titleIcon={resourceIconMapping.deployment_snapshot}>
            <Table
                isLoading={deploymentSnapshotsInfo.isLoading}
                columns={['ID', t('type'), t('bento version'), t('creator'), t('created_at')]}
                data={
                    deploymentSnapshotsInfo.data?.items.map((deploymentSnapshot) => [
                        deploymentSnapshot.uid,
                        <StyledLink
                            style={{
                                cursor: 'pointer',
                            }}
                            key={deploymentSnapshot.uid}
                            onClick={() => setDesiredShowDeploymentSnapshot(deploymentSnapshot)}
                        >
                            {t(deploymentSnapshot.type)}
                        </StyledLink>,
                        <Link
                            key={deploymentSnapshot.uid}
                            to={`/bentos/${deploymentSnapshot.bento_version.bento.name}/versions/${deploymentSnapshot.bento_version.version}`}
                        >
                            {deploymentSnapshot.bento_version.bento.name}:{deploymentSnapshot.bento_version.version}
                        </Link>,
                        deploymentSnapshot.creator && <User user={deploymentSnapshot.creator} />,
                        formatTime(deploymentSnapshot.created_at),
                    ]) ?? []
                }
                overrides={{
                    TableBodyRow: {
                        style: ({ $theme, $rowIndex }: { $theme: Theme; $rowIndex: number }) => {
                            const deploymentSnapshot = deploymentSnapshotsInfo.data?.items[$rowIndex]
                            if (!deploymentSnapshot) {
                                return {}
                            }
                            if (deploymentSnapshot.status !== 'active') {
                                return {}
                            }
                            const color_ =
                                deploymentSnapshot.type === 'stable'
                                    ? color($theme.colors.backgroundLightAccent).alpha(0.3).toString()
                                    : color($theme.colors.backgroundLightNegative).alpha(0.3).toString()
                            return {
                                'backgroundColor': color_,
                                ':hover': {
                                    backgroundColor: color_,
                                },
                            }
                        },
                    },
                }}
                paginationProps={{
                    start: deploymentSnapshotsInfo.data?.start,
                    count: deploymentSnapshotsInfo.data?.count,
                    total: deploymentSnapshotsInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        deploymentSnapshotsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateDeploymentSnapshotOpen}
                onClose={() => setIsCreateDeploymentSnapshotOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('deployment snapshot')])}</ModalHeader>
                <ModalBody>
                    <DeploymentForm
                        clusterName={clusterName}
                        deployment={deployment}
                        deploymentSnapshot={deployment?.latest_snapshot}
                        onSubmit={handleCreateDeploymentSnapshot}
                    />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={desiredShowDeploymentSnapshot !== undefined}
                onClose={() => setDesiredShowDeploymentSnapshot(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalBody>
                    {desiredShowDeploymentSnapshot && (
                        <DeploymentSnapshotDetail deploymentSnapshot={desiredShowDeploymentSnapshot} />
                    )}
                </ModalBody>
            </Modal>
        </Card>
    )
}
