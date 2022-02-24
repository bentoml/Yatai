import { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { updateDeployment } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import DeploymentForm from '@/components/DeploymentForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'
import { IUpdateDeploymentSchema } from '@/schemas/deployment'
import { useDeployment } from '@/hooks/useDeployment'
import { Theme } from 'baseui/theme'
import color from 'color'
import { useFetchDeploymentRevisions } from '@/hooks/useFetchDeploymentRevisions'
import { IDeploymentRevisionSchema } from '@/schemas/deployment_revision'
import { Button } from 'baseui/button'
import { useHistory } from 'react-router-dom'
import DeploymentTargetInfo from './DeploymentTargetInfo'

export interface IDeploymentRevisionListCardProps {
    clusterName: string
    kubeNamespace: string
    deploymentName: string
}

export default function DeploymentRevisionListCard({
    clusterName,
    kubeNamespace,
    deploymentName,
}: IDeploymentRevisionListCardProps) {
    const [page] = usePage()
    const { deployment } = useDeployment()
    const { deploymentRevisionsInfo } = useFetchDeploymentRevisions(clusterName, kubeNamespace, deploymentName, page)
    const [wishToDeployRevision, setWishToDeployRevision] = useState<IDeploymentRevisionSchema>()
    const [isCreateDeploymentRevisionOpen, setIsCreateDeploymentRevisionOpen] = useState(false)
    const history = useHistory()
    const handleCreateDeploymentRevision = useCallback(
        async (data: IUpdateDeploymentSchema) => {
            await updateDeployment(clusterName, kubeNamespace, deploymentName, data)
            await deploymentRevisionsInfo.refetch()
            setIsCreateDeploymentRevisionOpen(false)
            setWishToDeployRevision(undefined)
        },
        [clusterName, kubeNamespace, deploymentName, deploymentRevisionsInfo]
    )

    const [t] = useTranslation()

    return (
        <Card title={t('revisions')} titleIcon={resourceIconMapping.deployment_revision}>
            <Table
                isLoading={deploymentRevisionsInfo.isLoading}
                columns={['ID', t('deployment targets'), t('creator'), t('created_at'), t('operation')]}
                data={
                    deploymentRevisionsInfo.data?.items.map((deploymentRevision) => [
                        deploymentRevision.uid,
                        <div key={deploymentRevision.uid}>
                            {deploymentRevision.targets.map((target) => (
                                <DeploymentTargetInfo key={target.uid} deploymentTarget={target} />
                            ))}
                        </div>,
                        deploymentRevision.creator && <User user={deploymentRevision.creator} />,
                        formatDateTime(deploymentRevision.created_at),
                        <div key={deploymentRevision.uid}>
                            {deploymentRevision.status !== 'active' && (
                                <Button
                                    size='mini'
                                    onClick={() =>
                                        history.push(
                                            `/clusters/${clusterName}/namespaces/${kubeNamespace}/deployments/${deploymentName}/revisions/${deploymentRevision.uid}/rollback`
                                        )
                                    }
                                >
                                    {t('rollback')}
                                </Button>
                            )}
                        </div>,
                    ]) ?? []
                }
                overrides={{
                    TableBodyRow: {
                        style: ({ $theme, $rowIndex }: { $theme: Theme; $rowIndex: number }) => {
                            const deploymentRevision = deploymentRevisionsInfo.data?.items[$rowIndex]
                            if (!deploymentRevision) {
                                return {}
                            }
                            if (deploymentRevision.status !== 'active') {
                                return {}
                            }
                            const color_ = deploymentRevision.targets.find((target) => target.type === 'stable')
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
                    start: deploymentRevisionsInfo.data?.start,
                    count: deploymentRevisionsInfo.data?.count,
                    total: deploymentRevisionsInfo.data?.total,
                    afterPageChange: () => {
                        deploymentRevisionsInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateDeploymentRevisionOpen}
                onClose={() => setIsCreateDeploymentRevisionOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('update sth', [t('deployment')])}</ModalHeader>
                <ModalBody>
                    <DeploymentForm
                        clusterName={clusterName}
                        deployment={deployment}
                        deploymentRevision={deployment?.latest_revision}
                        onSubmit={handleCreateDeploymentRevision}
                    />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={wishToDeployRevision !== undefined}
                onClose={() => setWishToDeployRevision(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('update sth', [t('deployment')])}</ModalHeader>
                <ModalBody>
                    <DeploymentForm
                        clusterName={clusterName}
                        deployment={deployment}
                        deploymentRevision={wishToDeployRevision}
                        onSubmit={handleCreateDeploymentRevision}
                    />
                </ModalBody>
            </Modal>
        </Card>
    )
}
