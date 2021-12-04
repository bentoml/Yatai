import useTranslation from '@/hooks/useTranslation'
import { DeploymentTargetType, IDeploymentTargetSchema } from '@/schemas/deployment_target'
import { Modal, ModalBody } from 'baseui/modal'
import { Tag, KIND as TagKind } from 'baseui/tag'
import React, { useState } from 'react'
import { Link } from 'react-router-dom'
import DeploymentTargetDetail from './DeploymentTargetDetail'

const deploymentTargetTypeColorMap: Record<DeploymentTargetType, keyof TagKind> = {
    stable: TagKind.primary,
    canary: TagKind.accent,
}

export interface IDeploymentTargetInfoProps {
    deploymentTarget: IDeploymentTargetSchema
}

export default function DeploymentTargetInfo({ deploymentTarget }: IDeploymentTargetInfoProps) {
    const [showDetail, setShowDetail] = useState(false)
    const [t] = useTranslation()
    return (
        <>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 10,
                }}
            >
                <Tag
                    closeable={false}
                    variant='solid'
                    kind={deploymentTargetTypeColorMap[deploymentTarget.type]}
                    onClick={() => setShowDetail(true)}
                >
                    {t(deploymentTarget.type)}
                </Tag>
                <Link
                    to={`/bento_repositories/${deploymentTarget.bento.repository.name}/bentos/${deploymentTarget.bento.version}`}
                >
                    {deploymentTarget.bento.repository.name}:{deploymentTarget.bento.version}
                </Link>
            </div>
            <Modal isOpen={showDetail} onClose={() => setShowDetail(false)} closeable animate autoFocus>
                <ModalBody>
                    <DeploymentTargetDetail deploymentTarget={deploymentTarget} />
                </ModalBody>
            </Modal>
        </>
    )
}
