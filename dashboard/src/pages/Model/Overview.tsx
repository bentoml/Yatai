import React, { useState } from 'react'
import { useModel, useModelLoading } from '@/hooks/useModel'
import { Skeleton } from 'baseui/skeleton'
import { createUseStyles } from 'react-jss'
import useTranslation from '@/hooks/useTranslation'
import ImageBuildStatusTag from '@/components/ImageBuildStatusTag'
import { listModelBentos, listModelDeployments, recreateModelImageBuilderJob, updateModel } from '@/services/model'
import LabelList from '@/components/LabelList'
import Card from '@/components/Card'
import Time from '@/components/Time'
import User from '@/components/User'
import classNames from 'classnames'
import prettyBytes from 'pretty-bytes'
import { Link, useParams } from 'react-router-dom'
import { useFetchModel } from '@/hooks/useFetchModel'
import { resourceIconMapping } from '@/consts'
import { useQuery } from 'react-query'
import { IListQuerySchema } from '@/schemas/list'
import qs from 'qs'
import BentoList from '@/components/BentoList'
import { AiOutlineTags } from 'react-icons/ai'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { TiClipboard } from 'react-icons/ti'
import { Button } from 'baseui/button'
import CopyToClipboard from 'react-copy-to-clipboard'
import { docco, dark } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { Notification } from 'baseui/notification'
import List from '@/components/List'
import { IDeploymentSchema } from '@/schemas/deployment'
import DeploymentStatusTag from '@/components/DeploymentStatusTag'
import { useStyletron } from 'baseui'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'

const useStyles = createUseStyles({
    left: {
        flexGrow: 1,
        flexShrink: 0,
    },
    right: {
        flexGrow: 1,
        flexShrink: 0,
    },
    itemsWrapper: {
        display: 'flex',
        flexDirection: 'column',
        gap: 10,
    },
    item: {
        display: 'flex',
        alignItems: 'center',
        gap: 12,
    },
    key: {
        'flexShrink': 0,
        'display': 'flex',
        'alignItems': 'center',
        'fontWeight': 500,
        'gap': 6,
        '&:after': {
            content: '":"',
        },
    },
    value: {
        width: '100%',
    },
    foldedItem: {
        'alignItems': 'flex-start !important',
        '& $key': {
            cursor: 'pointer',
        },
    },
    closedItem: {
        '& > $key': {
            '&:before': {
                content: '"▲"',
            },
        },
    },
    openedItem: {
        '& > $key': {
            '&:before': {
                content: '"▼"',
            },
        },
    },
})

export default function ModelOverview() {
    const styles = useStyles()
    const { modelRepositoryName, modelVersion } = useParams<{ modelRepositoryName: string; modelVersion: string }>()
    const modelInfo = useFetchModel(modelRepositoryName, modelVersion)
    const { model } = useModel()
    const { modelLoading } = useModelLoading()
    const [t] = useTranslation()
    const [showContext, setShowContext] = useState(false)
    const [showMetaData, setShowMetaData] = useState(false)
    const [bentosQuery, setBentosQuery] = useState<IListQuerySchema>({
        start: 0,
        count: 10,
    })
    const bentosQueryKey = `model:${modelRepositoryName}/${modelVersion}:bentos:${qs.stringify(bentosQuery)}`
    const bentosInfo = useQuery(bentosQueryKey, () => listModelBentos(modelRepositoryName, modelVersion, bentosQuery))
    const [deploymentsQuery, setDeploymentsQuery] = useState<IListQuerySchema>({
        start: 0,
        count: 10,
    })
    const deploymentsQueryKey = `model:${modelRepositoryName}/${modelVersion}:deployments:${qs.stringify(
        deploymentsQuery
    )}`
    const deploymentsInfo = useQuery(deploymentsQueryKey, () =>
        listModelDeployments(modelRepositoryName, modelVersion, deploymentsQuery)
    )
    const downloadCommand = `bentoml models pull ${modelRepositoryName}:${modelVersion}`
    const [copyNotification, setCopyNotification] = useState<string>()
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const highlightTheme = themeType === 'dark' ? dark : docco

    if (modelLoading || !model) {
        return <Skeleton rows={3} animation />
    }

    return (
        <div
            style={{
                display: 'grid',
                gap: 20,
                gridTemplateColumns: '1fr 1fr',
            }}
        >
            <div className={styles.left}>
                <Card>
                    <div className={styles.itemsWrapper}>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('image build status')}</div>
                            <div className={styles.value}>
                                <ImageBuildStatusTag
                                    status={model.image_build_status}
                                    podsSelector={`yatai.io/model=${model.version},yatai.io/model-repository=${model.repository.name}`}
                                    onRerunClick={async () => {
                                        await recreateModelImageBuilderJob(model.repository.name, model.version)
                                    }}
                                />
                            </div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('created_at')}</div>
                            <div className={styles.value}>
                                <Time time={model.created_at} />
                            </div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('user')}</div>
                            <div className={styles.value}>{model.creator ? <User user={model.creator} /> : '-'}</div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('module')}</div>
                            <div className={styles.value}>{model.manifest.module}</div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>BentoML Version</div>
                            <div className={styles.value}>{model.manifest.bentoml_version}</div>
                        </div>
                        <div
                            className={classNames({
                                [styles.item]: true,
                                [styles.foldedItem]: true,
                                [styles.closedItem]: !showContext,
                                [styles.openedItem]: showContext,
                            })}
                        >
                            <div
                                className={styles.key}
                                onClick={() => setShowContext((v) => !v)}
                                role='button'
                                tabIndex={0}
                            >
                                {t('context')}
                            </div>
                            <div className={styles.value}>
                                {showContext ? (
                                    <div>
                                        {Object.entries(model.manifest.context).map(([key, value]) => (
                                            <div key={key} className={styles.item}>
                                                <div className={styles.key}>{key}</div>
                                                <div className={styles.value}>{value}</div>
                                            </div>
                                        ))}
                                    </div>
                                ) : (
                                    '...'
                                )}
                            </div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>Size</div>
                            <div className={styles.value}>{prettyBytes(model.manifest.size_bytes)}</div>
                        </div>
                        <div
                            className={styles.item}
                            style={{
                                alignItems: 'flex-start',
                            }}
                        >
                            <div className={styles.key}>{t('download')}</div>
                            <div className={styles.value}>
                                <div
                                    style={{
                                        display: 'flex',
                                        alignItems: 'flex-start',
                                        gap: 10,
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
                                            {downloadCommand}
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
                                    <div>
                                        <CopyToClipboard
                                            text={downloadCommand}
                                            onCopy={() => {
                                                setCopyNotification(t('copied to clipboard'))
                                            }}
                                        >
                                            <Button
                                                startEnhancer={<TiClipboard size={14} />}
                                                kind='secondary'
                                                size='compact'
                                            >
                                                {t('copy')}
                                            </Button>
                                        </CopyToClipboard>
                                    </div>
                                </div>
                            </div>
                        </div>
                        <div
                            className={classNames({
                                [styles.item]: true,
                                [styles.foldedItem]: true,
                                [styles.closedItem]: !showMetaData,
                                [styles.openedItem]: showMetaData,
                            })}
                        >
                            <div
                                className={styles.key}
                                role='button'
                                tabIndex={0}
                                onClick={() => setShowMetaData((v) => !v)}
                            >
                                Meta Data
                            </div>
                            <div className={styles.value}>
                                {showMetaData ? (
                                    <SyntaxHighlighter language='json' style={highlightTheme}>
                                        {JSON.stringify(model.manifest.metadata, null, 2)}
                                    </SyntaxHighlighter>
                                ) : (
                                    '...'
                                )}
                            </div>
                        </div>
                    </div>
                </Card>
            </div>
            <div className={styles.right}>
                <Card title={t('labels')} titleIcon={AiOutlineTags}>
                    <LabelList
                        value={model.labels}
                        onChange={async (labels) => {
                            await updateModel(model.repository.name, model.version, {
                                ...model,
                                labels,
                            })
                            await modelInfo.refetch()
                        }}
                    />
                </Card>
                <Card title={t('bentos')} titleIcon={resourceIconMapping.bento}>
                    <BentoList
                        isLoading={bentosInfo.isLoading}
                        bentos={bentosInfo.data?.items ?? []}
                        queryKey={bentosQueryKey}
                        paginationProps={{
                            start: bentosQuery.start,
                            count: bentosQuery.count,
                            total: bentosInfo.data?.total ?? 0,
                            onPageChange: (page) => {
                                setBentosQuery({
                                    ...bentosQuery,
                                    start: (page - 1) * bentosQuery.count,
                                })
                            },
                        }}
                    />
                </Card>
                <Card title={t('deployments')} titleIcon={resourceIconMapping.deployment}>
                    <List
                        isLoading={deploymentsInfo.isLoading}
                        items={deploymentsInfo.data?.items ?? []}
                        paginationProps={{
                            start: deploymentsQuery.start,
                            count: deploymentsQuery.count,
                            total: deploymentsInfo.data?.total ?? 0,
                            onPageChange: (page) => {
                                setDeploymentsQuery({
                                    ...deploymentsQuery,
                                    start: (page - 1) * deploymentsQuery.count,
                                })
                            },
                        }}
                        onRenderItem={(item: IDeploymentSchema) => {
                            return (
                                <div
                                    style={{
                                        padding: '6px 0',
                                        borderBottom: `1px solid ${theme.borders.border100.borderColor}`,
                                    }}
                                >
                                    <div
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            width: '100%',
                                            gap: 10,
                                        }}
                                    >
                                        <DeploymentStatusTag size='small' status={item.status} />
                                        <div
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                justifyContent: 'space-between',
                                                flexGrow: 1,
                                            }}
                                        >
                                            <Link to={`/clusters/${item.cluster?.name}/deployments/${item.name}`}>
                                                {item.name}
                                            </Link>
                                            <Time time={item.created_at} />
                                        </div>
                                    </div>
                                </div>
                            )
                        }}
                    />
                </Card>
            </div>
        </div>
    )
}
