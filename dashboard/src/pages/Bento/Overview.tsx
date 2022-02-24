import React, { useState } from 'react'
import { useBento, useBentoLoading } from '@/hooks/useBento'
import { Skeleton } from 'baseui/skeleton'
import { createUseStyles } from 'react-jss'
import useTranslation from '@/hooks/useTranslation'
import ImageBuildStatusTag from '@/components/ImageBuildStatusTag'
import { listBentoModels, listBentoDeployments, recreateBentoImageBuilderJob, updateBento } from '@/services/bento'
import LabelList from '@/components/LabelList'
import Card from '@/components/Card'
import Time from '@/components/Time'
import User from '@/components/User'
import prettyBytes from 'pretty-bytes'
import { useParams } from 'react-router-dom'
import { useFetchBento } from '@/hooks/useFetchBento'
import { resourceIconMapping } from '@/consts'
import { useQuery } from 'react-query'
import { IListQuerySchema } from '@/schemas/list'
import qs from 'qs'
import { AiOutlineCloudDownload, AiOutlineTags } from 'react-icons/ai'
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
import ModelList from '@/components/ModelList'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import Link from '@/components/Link'

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
    deploymentItem: (props: IThemedStyleProps) => ({
        'padding': '6px 2px',
        'cursor': 'pointer',
        'display': 'flex',
        'alignItems': 'center',
        'gap': 12,
        'borderBottom': `1px solid ${props.theme.borders.border100.borderColor}`,
        '&:hover': {
            backgroundColor: props.theme.colors.backgroundSecondary,
        },
    }),
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

export default function BentoOverview() {
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const styles = useStyles({ theme, themeType })
    const { bentoRepositoryName, bentoVersion } = useParams<{ bentoRepositoryName: string; bentoVersion: string }>()
    const bentoInfo = useFetchBento(bentoRepositoryName, bentoVersion)
    const { bento } = useBento()
    const { bentoLoading } = useBentoLoading()
    const [t] = useTranslation()
    const modelsQueryKey = `bento:${bentoRepositoryName}/${bentoVersion}:models`
    const modelsInfo = useQuery(modelsQueryKey, () => listBentoModels(bentoRepositoryName, bentoVersion))
    const [deploymentsQuery, setDeploymentsQuery] = useState<IListQuerySchema>({
        start: 0,
        count: 10,
    })
    const deploymentsQueryKey = `bento:${bentoRepositoryName}/${bentoVersion}:deployments:${qs.stringify(
        deploymentsQuery
    )}`
    const deploymentsInfo = useQuery(deploymentsQueryKey, () =>
        listBentoDeployments(bentoRepositoryName, bentoVersion, deploymentsQuery)
    )
    const downloadCommand = `bentoml pull ${bentoRepositoryName}:${bentoVersion}`
    const [copyNotification, setCopyNotification] = useState<string>()
    const highlightTheme = themeType === 'dark' ? dark : docco

    if (bentoLoading || !bento) {
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
                            <div className={styles.key}>{t('status')}</div>
                            <div className={styles.value}>
                                <ImageBuildStatusTag
                                    status={bento.image_build_status}
                                    podsSelector={`yatai.io/bento=${bento.version},yatai.io/bento-repository=${bento.repository.name}`}
                                    onRerunClick={async () => {
                                        await recreateBentoImageBuilderJob(bento.repository.name, bento.version)
                                    }}
                                />
                            </div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('created_at')}</div>
                            <div className={styles.value}>
                                <Time time={bento.created_at} />
                            </div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>{t('user')}</div>
                            <div className={styles.value}>{bento.creator ? <User user={bento.creator} /> : '-'}</div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>BentoML Version</div>
                            <div className={styles.value}>{bento.manifest.bentoml_version}</div>
                        </div>
                        <div className={styles.item}>
                            <div className={styles.key}>Size</div>
                            <div className={styles.value}>{prettyBytes(bento.manifest.size_bytes)}</div>
                        </div>
                    </div>
                </Card>
            </div>
            <div className={styles.right}>
                <Card title={t('download')} titleIcon={AiOutlineCloudDownload}>
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
                        <div style={{ flexShrink: 0 }}>
                            <CopyToClipboard
                                text={downloadCommand}
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
                </Card>
                <Card title={t('labels')} titleIcon={AiOutlineTags}>
                    <LabelList
                        value={bento.labels}
                        onChange={async (labels) => {
                            await updateBento(bento.repository.name, bento.version, {
                                ...bento,
                                labels,
                            })
                            await bentoInfo.refetch()
                        }}
                    />
                </Card>
                <Card title={t('models')} titleIcon={resourceIconMapping.bento}>
                    <ModelList
                        isLoading={modelsInfo.isLoading}
                        models={modelsInfo.data ?? []}
                        queryKey={modelsQueryKey}
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
                                    className={styles.deploymentItem}
                                    onClick={(e: React.MouseEvent) => {
                                        e.currentTarget.querySelector('a')?.click()
                                    }}
                                    role='button'
                                    tabIndex={0}
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
                                            <Link
                                                href={`/clusters/${item.cluster?.name}/namespaces/${item.kube_namespace}/deployments/${item.name}`}
                                            >
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
