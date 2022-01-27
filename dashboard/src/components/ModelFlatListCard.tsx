import { useCallback, useEffect, useMemo, useState } from 'react'
import { useQuery, useQueryClient } from 'react-query'
import Card from '@/components/Card'
import { listAllModels, recreateModelImageBuilderJob } from '@/services/model'
import { usePage } from '@/hooks/usePage'
import { IModelSchema, IModelWithRepositorySchema } from '@/schemas/model'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import { resourceIconMapping } from '@/consts'
import { useSubscription } from '@/hooks/useSubscription'
import { IListSchema } from '@/schemas/list'
import qs from 'qs'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { useFetchOrganizationModelModules } from '@/hooks/useFetchOrganizationModelModules'
import { useQ } from '@/hooks/useQ'
import prettyBytes from 'pretty-bytes'
import { MonoParagraphXSmall } from 'baseui/typography'
import { useHistory } from 'react-router-dom'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import SyntaxHighlighter from 'react-syntax-highlighter'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { dark, docco } from 'react-syntax-highlighter/dist/esm/styles/hljs'
import { Button } from 'baseui/button'
import FilterInput from './FilterInput'
import FilterBar from './FilterBar'
import { ResourceLabels } from './ResourceLabels'
import List from './List'
import ImageBuildStatusIcon from './ImageBuildStatusIcon'
import Time from './Time'
import Link from './Link'
import ListItem from './ListItem'

export default function ModelFlatListCard() {
    const { q, updateQ } = useQ()
    const [page] = usePage()
    const queryKey = `fetchAllModels:${qs.stringify(page)}`
    const modelsInfo = useQuery(queryKey, () => listAllModels(page))
    const membersInfo = useFetchOrganizationMembers()
    const modelModulesInfo = useFetchOrganizationModelModules()
    const [isCreateModelOpen, setIsCreateModelOpen] = useState(false)
    const [t] = useTranslation()
    const themeType = useCurrentThemeType()
    const highlightTheme = themeType === 'dark' ? dark : docco

    const uids = useMemo(
        () => modelsInfo.data?.items.map((modelVersion) => modelVersion.uid) ?? [],
        [modelsInfo.data?.items]
    )
    const queryClient = useQueryClient()
    const subscribeCb = useCallback(
        (modelVersion: IModelSchema) => {
            queryClient.setQueryData(queryKey, (oldData?: IListSchema<IModelSchema>): IListSchema<IModelSchema> => {
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
                    items: oldData.items.map((oldModelVersion) => {
                        if (oldModelVersion.uid === modelVersion.uid) {
                            return {
                                ...oldModelVersion,
                                ...modelVersion,
                            }
                        }
                        return oldModelVersion
                    }),
                }
            })
        },
        [queryClient, queryKey]
    )
    const { subscribe, unsubscribe } = useSubscription()

    useEffect(() => {
        subscribe({
            resourceType: 'model',
            resourceUids: uids,
            cb: subscribeCb,
        })
        return () => {
            unsubscribe({
                resourceType: 'model',
                resourceUids: uids,
                cb: subscribeCb,
            })
        }
    }, [subscribe, subscribeCb, uids, unsubscribe])

    const history = useHistory()

    const handleRenderItem = useCallback(
        (model: IModelWithRepositorySchema) => {
            return (
                <ListItem
                    onClick={() => {
                        history.push(`/model_repositories/${model.repository.name}/models/${model.version}`)
                    }}
                    key={model.uid}
                    artwork={() => (
                        <ImageBuildStatusIcon
                            key={model.uid}
                            status={model.image_build_status}
                            podsSelector={`yatai.io/model=${model.version},yatai.io/model-repository=${model.repository.name}`}
                            onRerunClick={async () => {
                                await recreateModelImageBuilderJob(model.repository.name, model.version)
                            }}
                        />
                    )}
                    endEnhancer={() => (
                        <div
                            style={{
                                display: 'flex',
                                flexDirection: 'column',
                                gap: 8,
                            }}
                        >
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 20,
                                }}
                            >
                                <div>{prettyBytes(model.manifest.size_bytes)}</div>
                                <div>{model.manifest.module}</div>
                            </div>
                            <div
                                style={{
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: 8,
                                }}
                            >
                                {model.creator && <User size='16px' user={model.creator} />}
                                {t('Created At')}
                                <Time time={model.created_at} />
                            </div>
                        </div>
                    )}
                >
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            gap: 10,
                        }}
                    >
                        <Link href={`/model_repositories/${model.repository.name}/models/${model.version}`}>
                            <MonoParagraphXSmall
                                overrides={{
                                    Block: {
                                        style: {
                                            margin: 0,
                                        },
                                    },
                                }}
                            >
                                {model.repository.name}:{model.version}
                            </MonoParagraphXSmall>
                        </Link>
                        <ResourceLabels resource={model} />
                    </div>
                </ListItem>
            )
        },
        [history, t]
    )

    return (
        <Card
            title={t('models')}
            titleIcon={resourceIconMapping.model}
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
                                    label: t('the models I created'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
            extra={
                <Button size='compact' onClick={() => setIsCreateModelOpen(true)}>
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
                            modelModulesInfo.data?.map((module) => ({
                                id: module,
                                label: module,
                            })) ?? [],
                        value: ((q.module as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            updateQ({
                                module: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('module'),
                    },
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
                        options: [
                            {
                                id: 'build_at-desc',
                                label: t('newest build'),
                            },
                            {
                                id: 'build_at-asc',
                                label: t('oldest build'),
                            },
                            {
                                id: 'size-desc',
                                label: t('largest'),
                            },
                            {
                                id: 'size-asc',
                                label: t('smallest'),
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
            <List
                isLoading={modelsInfo.isLoading}
                items={modelsInfo.data?.items ?? []}
                onRenderItem={handleRenderItem}
                paginationProps={{
                    start: modelsInfo.data?.start,
                    count: modelsInfo.data?.count,
                    total: modelsInfo.data?.total,
                    afterPageChange: () => {
                        modelsInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateModelOpen} onClose={() => setIsCreateModelOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('model')])}</ModalHeader>
                <ModalBody>
                    <div>
                        <p>
                            1. {t('Follow to [BentoML quickstart guide] to create your first Model. prefix')}
                            <Link
                                href='https://docs.bentoml.org/en/latest/quickstart.html#getting-started-page'
                                target='_blank'
                            >
                                {t('BentoML quickstart guide')}
                            </Link>
                            {t('Follow to [BentoML quickstart guide] to create your first Model. suffix')}
                        </p>
                        <p>
                            2. {t('Create an [API-token] and login your BentoML CLI. prefix')}
                            <Link href='/api_tokens' target='_blank'>
                                {t('api token')}
                            </Link>
                            {t('Create an [API-token] and login your BentoML CLI. suffix')}
                        </p>
                        <p>
                            3. {t('Push new Model to Yatai with the `bentoml models push` CLI command. prefix')}
                            <SyntaxHighlighter
                                language='bash'
                                style={highlightTheme}
                                customStyle={{
                                    margin: 0,
                                    display: 'inline',
                                    padding: 2,
                                }}
                            >
                                bentoml models push
                            </SyntaxHighlighter>
                            {t('Push new Model to Yatai with the `bentoml models push` CLI command. suffix')}
                        </p>
                    </div>
                </ModalBody>
            </Modal>
        </Card>
    )
}
