import { useState } from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import Card from '@/components/Card'
import { formatTime } from '@/utils/datetime'
import User from '@/components/User'
import { AiOutlineHistory } from 'react-icons/ai'
import { useParams } from 'react-router-dom'
import { StyledLink } from 'baseui/link'
import { Button } from 'baseui/button'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import DeploymentTerminalRecordList from '@/components/DeploymentTerminalRecordList'

export default function DeploymentOverview() {
    const { clusterName } = useParams<{ clusterName: string; deploymentName: string }>()
    const { deployment } = useDeployment()
    const { deploymentLoading } = useDeploymentLoading()

    const [t] = useTranslation()
    const [showTerminalRecordsModal, setShowTerminalRecordsModal] = useState(false)

    return (
        <div>
            <Card
                title={t('overview')}
                titleIcon={RiSurveyLine}
                extra={
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 10,
                        }}
                    >
                        <Button
                            size='mini'
                            startEnhancer={<AiOutlineHistory />}
                            onClick={() => setShowTerminalRecordsModal(true)}
                        >
                            {t('view terminal history')}
                        </Button>
                    </div>
                }
            >
                <Table
                    isLoading={deploymentLoading}
                    columns={[t('name'), 'URL', t('description'), t('creator'), t('created_at')]}
                    data={[
                        [
                            deployment?.name,
                            <div key={deployment?.uid}>
                                {deployment?.urls.map((url) => (
                                    <StyledLink key={url} href={url} target='_blank'>
                                        {url}
                                    </StyledLink>
                                ))}
                            </div>,
                            deployment?.description,
                            deployment?.creator && <User user={deployment?.creator} />,
                            deployment && formatTime(deployment.created_at),
                        ],
                    ]}
                />
            </Card>
            <Modal
                size='auto'
                isOpen={showTerminalRecordsModal}
                onClose={() => setShowTerminalRecordsModal(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('view terminal history')}</ModalHeader>
                <ModalBody style={{ flex: '1 1 0' }}>
                    {deployment && (
                        <DeploymentTerminalRecordList clusterName={clusterName} deploymentName={deployment.name} />
                    )}
                </ModalBody>
            </Modal>
        </div>
    )
}
