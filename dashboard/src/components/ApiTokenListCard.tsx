import React, { useCallback, useState } from 'react'
import Card from '@/components/Card'
import { createApiToken, deleteApiToken, updateApiToken } from '@/services/api_token'
import { usePage } from '@/hooks/usePage'
import { IApiTokenSchema, ICreateApiTokenSchema, IUpdateApiTokenSchema } from '@/schemas/api_token'
import ApiTokenForm from '@/components/ApiTokenForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { Modal, ModalHeader, ModalBody, ModalButton, ModalFooter } from 'baseui/modal'
import Table from '@/components/Table'
import { useFetchApiTokens } from '@/hooks/useFetchApiTokens'
import { resourceIconMapping } from '@/consts'
import { useStyletron } from 'baseui'
import { Input } from 'baseui/input'
import { TiClipboard } from 'react-icons/ti'
import { Notification } from 'baseui/notification'
import CopyToClipboard from 'react-copy-to-clipboard'
import { CopyBlock, solarizedDark, solarizedLight } from 'react-code-blocks'  // eslint-disable-line
import useGlobalState from '@/hooks/global'

export default function ApiTokenListCard() {
    const [page] = usePage()
    const apiTokensInfo = useFetchApiTokens(page)
    const [theTokenWishToShow, setTheTokenWishToShow] = useState<string>()
    const [theApiTokenWishToUpdate, setTheApiTokenWishToUpdate] = useState<IApiTokenSchema>()
    const [theApiTokenWishToDelete, setTheApiTokenWishToDelete] = useState<IApiTokenSchema>()
    const [isCreateApiTokenOpen, setIsCreateApiTokenOpen] = useState(false)
    const [copyNotification, setCopyNotification] = useState<string>()
    const [deleteApiTokenLoading, setDeleteApiTokenLoading] = useState(false)
    const [themeType] = useGlobalState('themeType')
    const handleCreateApiToken = useCallback(
        async (data: ICreateApiTokenSchema) => {
            const apiToken = await createApiToken(data)
            setCopyNotification(undefined)
            setTheTokenWishToShow(apiToken.token)
            await apiTokensInfo.refetch()
            setIsCreateApiTokenOpen(false)
        },
        [apiTokensInfo]
    )
    const handleUpdateApiToken = useCallback(
        async (data: IUpdateApiTokenSchema) => {
            if (theApiTokenWishToUpdate === undefined) {
                return
            }
            await updateApiToken(theApiTokenWishToUpdate.uid, data)
            await apiTokensInfo.refetch()
            setTheApiTokenWishToUpdate(undefined)
        },
        [apiTokensInfo, theApiTokenWishToUpdate]
    )
    const handleDeleteApiToken = useCallback(async () => {
        if (theApiTokenWishToDelete === undefined) {
            return
        }
        setDeleteApiTokenLoading(true)
        try {
            await deleteApiToken(theApiTokenWishToDelete.uid)
            await apiTokensInfo.refetch()
            setTheApiTokenWishToDelete(undefined)
        } finally {
            setDeleteApiTokenLoading(false)
        }
    }, [apiTokensInfo, theApiTokenWishToDelete])

    const [t] = useTranslation()
    const [, theme] = useStyletron()
    const copyCliCommand =
        `bentoml yatai login --api-token ${theTokenWishToShow} --endpoint ${window.location.origin}` ?? ''
    const codeTheme = themeType === 'light' ? solarizedLight : solarizedDark

    return (
        <Card
            title={t('sth list', [t('api token')])}
            titleIcon={resourceIconMapping.api_token}
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateApiTokenOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <Table
                isLoading={apiTokensInfo.isLoading}
                columns={[
                    t('name'),
                    t('scopes'),
                    t('description'),
                    t('last_used_at'),
                    t('expired_at'),
                    t('created_at'),
                    t('operation'),
                ]}
                data={
                    apiTokensInfo.data?.items.map((apiToken) => [
                        apiToken.name,
                        apiToken.scopes.join(', '),
                        apiToken.description,
                        apiToken.last_used_at ? formatDateTime(apiToken.last_used_at) : '-',
                        <span
                            key={apiToken.uid}
                            style={{
                                color: apiToken.is_expired ? theme.colors.negative : theme.colors.positive,
                            }}
                        >
                            {apiToken.expired_at ? formatDateTime(apiToken.expired_at) : '-'}
                        </span>,
                        formatDateTime(apiToken.created_at),
                        <div
                            key={apiToken.uid}
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 10,
                            }}
                        >
                            <Button size='mini' onClick={() => setTheApiTokenWishToUpdate(apiToken)}>
                                {t('update')}
                            </Button>
                            <Button
                                size='mini'
                                onClick={() => setTheApiTokenWishToDelete(apiToken)}
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
                                {t('delete')}
                            </Button>
                        </div>,
                    ]) ?? []
                }
                paginationProps={{
                    start: apiTokensInfo.data?.start,
                    count: apiTokensInfo.data?.count,
                    total: apiTokensInfo.data?.total,
                    afterPageChange: () => {
                        apiTokensInfo.refetch()
                    },
                }}
            />
            <Modal
                isOpen={isCreateApiTokenOpen}
                onClose={() => setIsCreateApiTokenOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('api token')])}</ModalHeader>
                <ModalBody>
                    <ApiTokenForm onSubmit={handleCreateApiToken} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={theApiTokenWishToUpdate !== undefined}
                onClose={() => setTheApiTokenWishToUpdate(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('update sth', [t('api token')])}</ModalHeader>
                <ModalBody>
                    {theApiTokenWishToUpdate && (
                        <ApiTokenForm apiToken={theApiTokenWishToUpdate} onSubmit={handleUpdateApiToken} />
                    )}
                </ModalBody>
            </Modal>
            <Modal
                isOpen={theApiTokenWishToDelete !== undefined}
                onClose={() => setTheApiTokenWishToDelete(undefined)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('are you sure to delete this api token?')}</ModalHeader>
                <ModalFooter>
                    <ModalButton size='compact' kind='tertiary' onClick={() => setTheApiTokenWishToDelete(undefined)}>
                        {t('cancel')}
                    </ModalButton>
                    <ModalButton
                        size='compact'
                        overrides={{
                            BaseButton: {
                                style: {
                                    background: theme.colors.negative,
                                },
                            },
                        }}
                        onClick={(e) => {
                            e.preventDefault()
                            if (!theApiTokenWishToDelete) {
                                return
                            }
                            handleDeleteApiToken()
                        }}
                        isLoading={deleteApiTokenLoading}
                    >
                        {t('ok')}
                    </ModalButton>
                </ModalFooter>
            </Modal>
            <Modal
                isOpen={!!theTokenWishToShow}
                onClose={() => setTheTokenWishToShow(undefined)}
                size='auto'
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('api token only show once time tips')}</ModalHeader>
                <ModalBody>
                    <div>
                        <div
                            style={{
                                display: 'flex',
                                gap: 10,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    flexDirection: 'column',
                                    gap: 4,
                                    flexGrow: 1,
                                }}
                            >
                                <Input value={theTokenWishToShow} disabled />
                                {copyNotification && (
                                    <Notification
                                        closeable
                                        onClose={() => setCopyNotification(undefined)}
                                        kind='positive'
                                        overrides={{
                                            Body: {
                                                style: {
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
                            <div>
                                <CopyToClipboard
                                    text={theTokenWishToShow ?? ''}
                                    onCopy={() => {
                                        setCopyNotification(t('copied to clipboard'))
                                    }}
                                >
                                    <Button startEnhancer={<TiClipboard size={14} />} kind='secondary'>
                                        {t('copy')}
                                    </Button>
                                </CopyToClipboard>
                            </div>
                        </div>
                        <div>
                            <p>{t('copy command to login yatai')}</p>
                            <CopyBlock
                                text={copyCliCommand}
                                language='shell'
                                showLineNumbers={false}
                                theme={codeTheme}
                                wrapLongLines
                                codeBlock
                            />
                        </div>
                    </div>
                </ModalBody>
            </Modal>
        </Card>
    )
}
