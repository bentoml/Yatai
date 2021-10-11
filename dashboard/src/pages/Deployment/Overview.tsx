import React, { useState } from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import { VscServerProcess } from 'react-icons/vsc'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import Card from '@/components/Card'
import { formatTime } from '@/utils/datetime'
import User from '@/components/User'
import { AiOutlineHistory, AiOutlineQuestionCircle } from 'react-icons/ai'
import { IKubePodSchema } from '@/schemas/kube_pod'
import { useFetchDeploymentPods } from '@/hooks/useFetchDeploymentPods'
import { useHistory, useParams } from 'react-router-dom'
import { StyledLink } from 'baseui/link'
import PodList from '@/components/PodList'
import { Button } from 'baseui/button'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { FaJournalWhills } from 'react-icons/fa'
import DeploymentTerminalRecordList from '@/components/DeploymentTerminalRecordList'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import { StatefulTooltip } from 'baseui/tooltip'

export default function DeploymentOverview() {
    const { orgName, clusterName, deploymentName } =
        useParams<{ orgName: string; clusterName: string; deploymentName: string }>()
    const { deployment } = useDeployment()
    const { deploymentLoading } = useDeploymentLoading()
    const [pods, setPods] = useState<IKubePodSchema[]>()
    const [podsLoading, setPodsLoading] = useState(false)
    const { yataiComponentsInfo } = useFetchYataiComponents(orgName, clusterName)

    const hasLogging = yataiComponentsInfo.data?.find((x) => x.type === 'logging') !== undefined

    useFetchDeploymentPods({
        orgName,
        clusterName,
        deploymentName,
        setPods,
        setPodsLoading,
    })

    const [t] = useTranslation()
    const [showTerminalRecordsModal, setShowTerminalRecordsModal] = useState(false)
    const history = useHistory()

    return (
        <>
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
                        {hasLogging ? (
                            <Button
                                size='mini'
                                startEnhancer={<FaJournalWhills />}
                                onClick={() => {
                                    history.push(
                                        `/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/log`
                                    )
                                }}
                            >
                                {t('view log')}
                            </Button>
                        ) : (
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 5,
                                }}
                            >
                                <Button
                                    disabled
                                    size='mini'
                                    startEnhancer={<FaJournalWhills />}
                                    onClick={() => {
                                        history.push(
                                            `/orgs/${orgName}/clusters/${clusterName}/deployments/${deploymentName}/log`
                                        )
                                    }}
                                >
                                    {t('view log')}
                                </Button>
                                <StatefulTooltip
                                    content={t('please install yatai component first', [t('logging')])}
                                    showArrow
                                >
                                    <div
                                        style={{
                                            cursor: 'pointer',
                                        }}
                                    >
                                        <AiOutlineQuestionCircle />
                                    </div>
                                </StatefulTooltip>
                            </div>
                        )}
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
            <Card title={t('replicas')} titleIcon={VscServerProcess}>
                <PodList loading={podsLoading} pods={pods ?? []} />
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
                        <DeploymentTerminalRecordList
                            orgName={orgName}
                            clusterName={clusterName}
                            deploymentName={deployment.name}
                        />
                    )}
                </ModalBody>
            </Modal>
        </>
    )
}
