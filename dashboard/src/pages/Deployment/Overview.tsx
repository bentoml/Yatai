import { useState } from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useDeployment, useDeploymentLoading } from '@/hooks/useDeployment'
import Card from '@/components/Card'
import { formatDateTime } from '@/utils/datetime'
import User from '@/components/User'
import { AiOutlineHistory } from 'react-icons/ai'
import { useParams } from 'react-router-dom'
import { Button } from 'baseui/button'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import DeploymentTerminalRecordList from '@/components/DeploymentTerminalRecordList'
import Link from '@/components/Link'
import { MdEventNote, MdOutlineAccessibilityNew } from 'react-icons/md'
import KubePodEvents from '@/components/KubePodEvents'
import CopyableText from '@/components/CopyableText'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { docco, dark } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { Notification } from 'baseui/notification'
import CopyToClipboard from 'react-copy-to-clipboard'
import { TiClipboard } from 'react-icons/ti'

export default function DeploymentOverview() {
    const { clusterName, kubeNamespace, deploymentName } =
        useParams<{ clusterName: string; kubeNamespace: string; deploymentName: string }>()
    const { deployment } = useDeployment()
    const { deploymentLoading } = useDeploymentLoading()

    const [t] = useTranslation()
    const [showTerminalRecordsModal, setShowTerminalRecordsModal] = useState(false)
    const [showAccessDeploymentModal, setShowAccessDeploymentModal] = useState(false)
    const themeType = useCurrentThemeType()
    const highlightTheme = themeType === 'dark' ? dark : docco
    const portForwardCommand = `kubectl port-forward -n ${kubeNamespace} svc/${deploymentName} 8080:3000`
    const [copyNotification, setCopyNotification] = useState<string>()

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
                        {(!deployment?.urls || deployment.urls.length === 0) && (
                            <Button
                                size='mini'
                                startEnhancer={<MdOutlineAccessibilityNew />}
                                onClick={() => setShowAccessDeploymentModal(true)}
                            >
                                {t('accessing deployments from outside the cluster')}
                            </Button>
                        )}
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
                    columns={[
                        t('name'),
                        'URL',
                        t('cluster internal url'),
                        t('kube namespace'),
                        t('description'),
                        t('creator'),
                        t('created_at'),
                    ]}
                    data={[
                        [
                            deployment?.name,
                            <div key={deployment?.uid}>
                                {deployment?.urls.map((url) => (
                                    <Link key={url} href={url} target='_blank'>
                                        {url}
                                    </Link>
                                ))}
                            </div>,
                            <CopyableText
                                key={deployment?.uid}
                                text={`http://${deployment?.name}.${deployment?.kube_namespace}.svc.cluster.local:3000`}
                            />,
                            // eslint-disable-next-line jsx-a11y/no-static-element-interactions
                            <span key={deployment?.uid} onClick={(e) => e.stopPropagation()} style={{ cursor: 'text' }}>
                                {deployment?.kube_namespace}
                            </span>,
                            deployment?.description,
                            deployment?.creator && <User user={deployment?.creator} />,
                            deployment && formatDateTime(deployment.created_at),
                        ],
                    ]}
                />
            </Card>
            <Card title={t('events')} titleIcon={MdEventNote}>
                <KubePodEvents
                    open
                    width='auto'
                    height={200}
                    clusterName={clusterName}
                    namespace={kubeNamespace}
                    deploymentName={deploymentName}
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
                        <DeploymentTerminalRecordList
                            clusterName={clusterName}
                            kubeNamespace={deployment.kube_namespace}
                            deploymentName={deployment.name}
                        />
                    )}
                </ModalBody>
            </Modal>
            <Modal
                size='auto'
                isOpen={showAccessDeploymentModal}
                onClose={() => setShowAccessDeploymentModal(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('accessing deployments from outside the cluster')}</ModalHeader>
                <ModalBody>
                    <div>
                        <p>{t('deployment accessing tips')}</p>
                        <p>{t('the first way of the deployment accessing')}</p>
                        <p>{t('the second way of the deployment accessing')}</p>
                        <div
                            style={{
                                display: 'flex',
                                alignItems: 'flex-start',
                                gap: 10,
                                marginLeft: 30,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    flexGrow: 1,
                                }}
                            >
                                <SyntaxHighlighter
                                    language='bash'
                                    style={highlightTheme}
                                    customStyle={{
                                        margin: 0,
                                    }}
                                >
                                    {portForwardCommand}
                                </SyntaxHighlighter>
                                {copyNotification && (
                                    <Notification
                                        closeable
                                        onClose={() => setCopyNotification(undefined)}
                                        kind='positive'
                                        overrides={{
                                            Body: {
                                                style: {
                                                    margin: 0,
                                                    width: '100%',
                                                    boxSizing: 'border-box',
                                                    padding: '8px !important',
                                                    borderRadius: '3px !important',
                                                    fontSize: '13px !important',
                                                },
                                            },
                                        }}
                                    >
                                        {copyNotification}
                                    </Notification>
                                )}
                            </div>
                            <div style={{ flexShrink: 0 }}>
                                <CopyToClipboard
                                    text={portForwardCommand}
                                    onCopy={() => {
                                        setCopyNotification(t('copied to clipboard'))
                                    }}
                                >
                                    <Button startEnhancer={<TiClipboard size={14} />} kind='secondary' size='compact'>
                                        {t('copy')}
                                    </Button>
                                </CopyToClipboard>
                            </div>
                        </div>
                    </div>
                </ModalBody>
            </Modal>
        </div>
    )
}
