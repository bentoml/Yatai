import React, { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listBentoRepositories } from '@/services/bento_repository'
import { usePage } from '@/hooks/usePage'
import { IBentoRepositoryWithLatestDeploymentsSchema } from '@/schemas/bento_repository'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import qs from 'qs'
import { IDeploymentSchema } from '@/schemas/deployment'
import { IBentoSchema } from '@/schemas/bento'
import { useQ } from '@/hooks/useQ'
import { LabelMedium, LabelXSmall, MonoParagraphXSmall } from 'baseui/typography'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { recreateBentoImageBuilderJob } from '@/services/bento'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { dark, docco } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import FilterBar from './FilterBar'
import FilterInput from './FilterInput'
import Time from './Time'
import Grid from './Grid'
import List from './List'
import DeploymentStatusTag from './DeploymentStatusTag'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import Link from './Link'

const useStyles = createUseStyles({
    item: (props: IThemedStyleProps) => ({
        'display': 'flex',
        'alignItems': 'center',
        'padding': '2px',
        'borderBottom': `1px solid ${props.theme.borders.border100.borderColor}`,
        'cursor': 'pointer',
        '&:hover': {
            backgroundColor: props.theme.colors.backgroundSecondary,
        },
    }),
    itemsContainer: () => ({
        '& $item:last-child': {
            borderBottom: 'none',
        },
    }),
})

export default function BentoRepositoryListCard() {
    const [, theme] = useStyletron()
    const themeType = useCurrentThemeType()
    const styles = useStyles({ theme, themeType })
    const { q, updateQ } = useQ()
    const membersInfo = useFetchOrganizationMembers()
    const [page] = usePage()
    const queryKey = `fetchBentoRepositories:${qs.stringify(page)}`
    const bentoRepositoriesInfo = useQuery(queryKey, () => listBentoRepositories(page))
    const [isCreateBentoOpen, setIsCreateBentoOpen] = useState(false)
    const [t] = useTranslation()
    const highlightTheme = themeType === 'dark' ? dark : docco

    const queryClient = useQueryClient()
    const bentoUids = useMemo(
        () =>
            bentoRepositoriesInfo.data?.items.reduce(
                (acc, cur) => [...acc, ...cur.latest_bentos.map((x) => x.uid)],
                [] as string[]
            ) ?? [],
        [bentoRepositoriesInfo.data?.items]
    )
    const subscribeBentoCb = useCallback(
        (bento: IBentoSchema) => {
            queryClient.setQueryData(
                queryKey,
                (
                    oldData?: IListSchema<IBentoRepositoryWithLatestDeploymentsSchema>
                ): IListSchema<IBentoRepositoryWithLatestDeploymentsSchema> => {
                    if (!oldData) {
                        return {
                            start: 0,
                            count: 0,
                            total: 0,
                            items: [],
                        }
                    }
                    return {
                        ...oldData,
                        items: oldData.items.map((oldBentoRepository) => {
                            return {
                                ...oldBentoRepository,
                                latest_bentos: oldBentoRepository.latest_bentos.map((oldBento) => {
                                    if (oldBento.uid === bento.uid) {
                                        return bento
                                    }
                                    return oldBento
                                }),
                            }
                        }),
                    }
                }
            )
        },
        [queryClient, queryKey]
    )
    const deploymentUids = useMemo(
        () =>
            bentoRepositoriesInfo.data?.items.reduce(
                (acc, cur) => [...acc, ...cur.latest_deployments.map((x) => x.uid)],
                [] as string[]
            ) ?? [],
        [bentoRepositoriesInfo.data?.items]
    )
    const subscribeDeploymentCb = useCallback(
        (deployment: IDeploymentSchema) => {
            queryClient.setQueryData(
                queryKey,
                (
                    oldData?: IListSchema<IBentoRepositoryWithLatestDeploymentsSchema>
                ): IListSchema<IBentoRepositoryWithLatestDeploymentsSchema> => {
                    if (!oldData) {
                        return {
                            start: 0,
                            count: 0,
                            total: 0,
                            items: [],
                        }
                    }
                    return {
                        ...oldData,
                        items: oldData.items.map((oldBentoRepository) => {
                            return {
                                ...oldBentoRepository,
                                latest_deployments: oldBentoRepository.latest_deployments.map((oldDeployment) => {
                                    if (oldDeployment.uid === deployment.uid) {
                                        return deployment
                                    }
                                    return oldDeployment
                                }),
                            }
                        }),
                    }
                }
            )
        },
        [queryClient, queryKey]
    )

    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'bento',
            resourceUids: bentoUids,
            cb: subscribeBentoCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'bento',
                resourceUids: bentoUids,
                cb: subscribeBentoCb,
            })
        }
    }, [subscribe, unsubscribe, bentoUids, subscribeBentoCb])

    useEffect(() => {
        subscribe({
            resourceType: 'deployment',
            resourceUids: deploymentUids,
            cb: subscribeDeploymentCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'deployment',
                resourceUids: deploymentUids,
                cb: subscribeDeploymentCb,
            })
        }
    }, [deploymentUids, subscribe, subscribeDeploymentCb, unsubscribe])

    const handleRenderItem = useCallback(
        (bentoRepository: IBentoRepositoryWithLatestDeploymentsSchema) => {
            return (
                <div
                    style={{
                        position: 'relative',
                        height: 'calc(100% - 40px)',
                    }}
                >
                    <div
                        style={{
                            position: 'relative',
                            display: 'flex',
                            alignItems: 'center',
                            justifyContent: 'center',
                            fontSize: '14px',
                            gap: 16,
                            paddingBottom: 30,
                        }}
                    >
                        <div
                            style={{
                                position: 'absolute',
                                left: 0,
                                top: 0,
                                display: 'flex',
                                alignItems: 'center',
                                gap: 4,
                            }}
                        >
                            <div
                                style={{
                                    display: 'inline-flex',
                                }}
                            >
                                {React.createElement(resourceIconMapping.bento, { size: 18 })}
                            </div>
                            <div>{bentoRepository.n_bentos}</div>
                        </div>
                        <LabelMedium>
                            <Link href={`/bento_repositories/${bentoRepository.name}`}>{bentoRepository.name}</Link>
                        </LabelMedium>
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 20,
                        }}
                    >
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 3,
                            }}
                        >
                            <div
                                style={{
                                    paddingBottom: 10,
                                    borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                                }}
                            >
                                <LabelXSmall>{t('latest deployments')}</LabelXSmall>
                            </div>
                            <List
                                emptyText={t('no deployment found')}
                                items={bentoRepository.latest_deployments}
                                itemsContainerClassName={styles.itemsContainer}
                                onRenderItem={(item: IDeploymentSchema) => {
                                    return (
                                        <div
                                            key={item.uid}
                                            className={styles.item}
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 10,
                                            }}
                                            onClick={(e: React.MouseEvent) => {
                                                e.currentTarget.querySelector('a')?.click()
                                            }}
                                            role='button'
                                            tabIndex={0}
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
                                    )
                                }}
                            />
                        </div>
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 3,
                            }}
                        >
                            <div
                                style={{
                                    paddingBottom: 10,
                                    borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                                }}
                            >
                                <LabelXSmall>{t('latest versions')}</LabelXSmall>
                            </div>
                            <List
                                items={bentoRepository.latest_bentos}
                                itemsContainerClassName={styles.itemsContainer}
                                onRenderItem={(item: IBentoSchema) => {
                                    return (
                                        <div
                                            className={styles.item}
                                            key={item.uid}
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 10,
                                            }}
                                            onClick={(e: React.MouseEvent) => {
                                                e.currentTarget.querySelector('a')?.click()
                                            }}
                                            role='button'
                                            tabIndex={0}
                                        >
                                            <ImageBuildStatusIcon
                                                size={14}
                                                status={item.image_build_status}
                                                podsSelector={`yatai.ai/bento=${item.version},yatai.ai/bento-repository=${bentoRepository.name}`}
                                                onRerunClick={async () => {
                                                    await recreateBentoImageBuilderJob(
                                                        bentoRepository.name,
                                                        item.version
                                                    )
                                                }}
                                            />
                                            <div
                                                style={{
                                                    display: 'flex',
                                                    flexGrow: 1,
                                                    justifyContent: 'space-between',
                                                }}
                                            >
                                                <Link
                                                    href={`/bento_repositories/${bentoRepository.name}/bentos/${item.version}`}
                                                >
                                                    <MonoParagraphXSmall
                                                        overrides={{
                                                            Block: {
                                                                style: {
                                                                    margin: 0,
                                                                },
                                                            },
                                                        }}
                                                    >
                                                        {item.version}
                                                    </MonoParagraphXSmall>
                                                </Link>
                                                <Time time={item.created_at} />
                                            </div>
                                        </div>
                                    )
                                }}
                            />
                        </div>
                    </div>
                </div>
            )
        },
        [styles.item, styles.itemsContainer, t, theme.borders.border200.borderColor]
    )

    return (
        <Card
            title={t('bento repositories')}
            titleIcon={resourceIconMapping.bento}
            middle={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexGrow: 1,
                    }}
                >
                    <div
                        style={{
                            width: 100,
                            flexGrow: 1,
                        }}
                    />
                    <div
                        style={{
                            flexGrow: 2,
                            flexShrink: 0,
                            maxWidth: 1200,
                        }}
                    >
                        <FilterInput
                            filterConditions={[
                                {
                                    qStr: 'creator:@me',
                                    label: t('the bentos I created'),
                                },
                                {
                                    qStr: 'last_updater:@me',
                                    label: t('my last updated bentos'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <FilterBar
                filters={[
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],
                        value: ((q.creator as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                creator: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('creator'),
                    },
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],
                        value: ((q.last_updater as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                last_updater: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('last updater'),
                    },
                    {
                        options: [
                            {
                                id: 'updated_at-desc',
                                label: t('newest update'),
                            },
                            {
                                id: 'updated_at-asc',
                                label: t('oldest update'),
                            },
                        ],
                        value: ((q.sort as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                sort: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('sort'),
                    },
                ]}
            />
            <Grid
                isLoading={bentoRepositoriesInfo.isLoading}
                items={bentoRepositoriesInfo.data?.items ?? []}
                onRenderItem={handleRenderItem}
                paginationProps={{
                    start: bentoRepositoriesInfo.data?.start,
                    count: bentoRepositoriesInfo.data?.count,
                    total: bentoRepositoriesInfo.data?.total,
                    afterPageChange: () => {
                        bentoRepositoriesInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateBentoOpen} onClose={() => setIsCreateBentoOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('bento')])}</ModalHeader>
                <ModalBody>
                    <div>
                        <p>
                            1. {t('Follow to [BentoML quickstart guide] to create your first Bento. prefix')}
                            <Link
                                href='https://docs.bentoml.org/en/latest/quickstart.html#getting-started-page'
                                target='_blank'
                            >
                                {t('BentoML quickstart guide')}
                            </Link>
                            {t('Follow to [BentoML quickstart guide] to create your first Bento. suffix')}
                        </p>
                        <p>
                            2. {t('Create an [API-token] and login your BentoML CLI. prefix')}
                            <Link href='/api_tokens' target='_blank'>
                                {t('api token')}
                            </Link>
                            {t('Create an [API-token] and login your BentoML CLI. suffix')}
                        </p>
                        <p>
                            3. {t('Push new Bento to Yatai with the `bentoml push` CLI command. prefix')}
                            <SyntaxHighlighter
                                language='bash'
                                style={highlightTheme}
                                customStyle={{
                                    margin: 0,
                                    display: 'inline',
                                    padding: 2,
                                }}
                            >
                                bentoml push
                            </SyntaxHighlighter>
                            {t('Push new Bento to Yatai with the `bentoml push` CLI command. suffix')}
                        </p>
                    </div>
                </ModalBody>
            </Modal>
        </Card>
    )
}
