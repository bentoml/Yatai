import { useDeployment } from '@/hooks/useDeployment'
import useTranslation, { Translator } from '@/hooks/useTranslation'
import { ILokiLabelFilterNode, ILokiLineFilterNode } from '@/interfaces/ILoki'
import { IDeploymentSchema } from '@/schemas/deployment'
import { useStyletron } from 'baseui'
import { FaJournalWhills } from 'react-icons/fa'
import { useCluster } from '@/hooks/useCluster'

import color from 'color'
import _ from 'lodash'
import React, { useCallback, useState } from 'react'
import { createUseStyles } from 'react-jss'
import { Button } from 'baseui/button'
import { Input } from 'baseui/input'
import { Search } from 'baseui/icon'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { useFetchOrganizationMajorCluster } from '@/hooks/useFetchOrganizationMajorCluster'
import { useFetchBentoOptional } from '@/hooks/useFetchBento'
import { Select } from 'baseui/select'
import { resourceIconMapping } from '@/consts'
import Card from './Card'
import GrafanaIFrame from './GrafanaIFrame'
import LokiFiltersForm from './LokiFiltersForm'
import Text from './Text'
import Label from './Label'

const labelFilterNodeToString = (filterNode: ILokiLabelFilterNode): string => {
    return `${filterNode.name}${filterNode.operator}"${filterNode.value}"`
}

const labelFilterNodesToString = (filterNodes: ILokiLabelFilterNode[]): string => {
    return filterNodes.map(labelFilterNodeToString).join(', ')
}

const lineFilterNodeToString = (filterNode: ILokiLineFilterNode): string => {
    const value = `"${_.replace(filterNode.value, /"/g, '\\"')}"`

    if (filterNode.type === 'contains') {
        if (filterNode.isRegexp) {
            return `|~ ${value}`
        }
        return `|= ${value}`
    }
    if (filterNode.isRegexp) {
        return `!~ ${value}`
    }
    return `!= ${value}`
}

const lineFilterNodesToReadableString = (t: Translator, filterNodes: ILokiLineFilterNode[]): string => {
    return filterNodes
        .map((filter) => {
            let verb = ''
            if (filter.type === 'contains') {
                if (filter.isRegexp) {
                    verb = t('match')
                } else {
                    verb = t('contains')
                }
            } else if (filter.isRegexp) {
                verb = t('not match')
            } else {
                verb = t('not contains')
            }
            return `${verb} ${filter.value} `
        })
        .join(` ${t('and')} `)
}

const useStyles = createUseStyles({
    wrapper: {
        width: '100%',
        height: '100%',
    },
    iframe: {
        border: 0,
        width: '100%',
        height: '100%',
    },
})

interface ILokiLogProps {
    deployment?: IDeploymentSchema
    podName?: string
    namespace?: string
    style?: React.CSSProperties
}

export default function LokiLog({ deployment: deployment_, podName, namespace, style }: ILokiLogProps) {
    const styles = useStyles()

    const [t] = useTranslation()

    const majorClusterInfo = useFetchOrganizationMajorCluster()
    let { cluster } = useCluster()
    if (cluster === undefined && majorClusterInfo !== undefined) {
        cluster = majorClusterInfo.data
    }
    const { deployment: deployment0 } = useDeployment()

    let deployment = deployment_
    if (!deployment) {
        deployment = deployment0
    }

    const bentoInfo = useFetchBentoOptional(
        deployment?.latest_revision?.targets?.[0]?.bento.repository.name,
        deployment?.latest_revision?.targets?.[0]?.bento.version
    )

    const [keyword, setKeyword] = useState<string | undefined>()
    const [keyword_, setKeyword_] = useState<string | undefined>()

    const [labelFilterNodes, setLabelFilterNodes] = useState<ILokiLabelFilterNode[]>([])
    const [lineFilterNodes, setLineFilterNodes] = useState<ILokiLineFilterNode[]>([])
    const [openLineFiltersModal, setOpenLineFiltersModal] = useState(false)

    const onLineFilterNodesSubmit = useCallback((filterNodes) => {
        setLineFilterNodes(filterNodes)
        setOpenLineFiltersModal(false)
    }, [])

    const [, theme] = useStyletron()

    const [tempMaxLines, setTempMaxLines] = useState(1000)
    const [maxLines, setMaxLines] = useState(tempMaxLines)

    const grafanaRootPath = cluster?.grafana_root_path
    const dataSource = 'Loki'

    if (!grafanaRootPath) {
        return <div>no data</div>
    }

    let defaultLabelFilterNodes: ILokiLabelFilterNode[] = []
    let defaultLineFilterNodes: ILokiLineFilterNode[] = []
    if (deployment) {
        defaultLabelFilterNodes = [
            {
                name: 'yatai_ai_deployment',
                value: deployment.name,
                operator: '=',
            },
        ]
        defaultLineFilterNodes = [
            {
                type: 'not contains',
                value: 'kube-probe',
                isRegexp: false,
            },
        ]
    }

    if (podName) {
        defaultLabelFilterNodes = [
            ...defaultLabelFilterNodes,
            {
                name: 'pod',
                value: podName,
                operator: '=',
            },
        ]
    }

    if (namespace) {
        defaultLabelFilterNodes = [
            ...defaultLabelFilterNodes,
            {
                name: 'namespace',
                value: namespace,
                operator: '=',
            },
        ]
    }

    if (keyword) {
        defaultLineFilterNodes = [
            ...defaultLineFilterNodes,
            {
                type: 'contains',
                value: `(?i)${keyword}`,
                isRegexp: true,
            },
        ]
    }

    const expr = `{${labelFilterNodesToString(defaultLabelFilterNodes)}} ${
        labelFilterNodes.length > 0 ? `| ${labelFilterNodes.map(labelFilterNodeToString).join(' or ')}` : ''
    } ${[...defaultLineFilterNodes, ...lineFilterNodes].map(lineFilterNodeToString).join(' ')}`

    const pathname = `${grafanaRootPath}explore`
    const query = {
        kiosk: null,
        orgId: 1,
        left: JSON.stringify([
            'now-1h',
            'now',
            dataSource,
            {
                expr,
                maxLines,
            },
        ]),
    }

    const filtersReadableString = lineFilterNodesToReadableString(t, lineFilterNodes)

    return (
        <Card
            title='Loki Log'
            titleIcon={FaJournalWhills}
            className={styles.wrapper}
            bodyStyle={{
                height: '100%',
            }}
            extra={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexDirection: 'row',
                        gap: 10,
                    }}
                >
                    <div
                        style={{
                            maxWidth: 900,
                            overflow: 'hidden',
                            textOverflow: 'ellipsis',
                            whiteSpace: 'nowrap',
                        }}
                    >
                        <Text
                            style={{
                                color: color(theme.colors.contentPrimary).lighten(0.3).fade(0.2).rgb().string(),
                            }}
                        >
                            {filtersReadableString}
                        </Text>
                    </div>
                    <div
                        style={{
                            flexShrink: 0,
                        }}
                    >
                        <Button size='mini' onClick={() => setOpenLineFiltersModal(true)}>
                            {t('advanced search')}
                        </Button>
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            flexShrink: 0,
                            gap: 10,
                        }}
                    >
                        <Label
                            style={{
                                flexShrink: 0,
                            }}
                        >
                            {t('component')}
                        </Label>
                        <Select
                            multi
                            size='mini'
                            overrides={{
                                Root: {
                                    style: {
                                        minWidth: '300px',
                                    },
                                },
                            }}
                            clearable={false}
                            searchable={false}
                            options={[
                                {
                                    id: 'yatai_ai_is_bento_api_server="true"',
                                    label: (
                                        <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                            {React.createElement(resourceIconMapping.bento_api_server, { size: 12 })}
                                            <span>API Server</span>
                                        </div>
                                    ),
                                    filterNode: {
                                        name: 'yatai_ai_is_bento_api_server',
                                        value: 'true',
                                        operator: '=',
                                    },
                                },
                                ...(bentoInfo?.data?.manifest?.runners?.map((runner) => {
                                    return {
                                        id: `yatai_ai_bento_runner="${runner.name}"`,
                                        label: (
                                            <div
                                                style={{
                                                    display: 'flex',
                                                    alignItems: 'center',
                                                    gap: 3,
                                                }}
                                            >
                                                <div style={{ display: 'flex', alignItems: 'center', gap: 8 }}>
                                                    {React.createElement(resourceIconMapping.bento_runner, {
                                                        size: 12,
                                                    })}
                                                    <span style={{ fontWeight: 'bold' }}>Runner</span>
                                                </div>
                                                {runner.name}
                                            </div>
                                        ),
                                        filterNode: {
                                            name: 'yatai_ai_bento_runner',
                                            value: runner.name,
                                            operator: '=',
                                        },
                                    }
                                }) ?? []),
                            ]}
                            value={labelFilterNodes.map((filterNode) => {
                                return {
                                    id: labelFilterNodeToString(filterNode),
                                    filterNode,
                                }
                            })}
                            onChange={(params) => {
                                // eslint-disable-next-line no-console
                                console.log(params)
                                setLabelFilterNodes(
                                    params.value.map((param) => param.filterNode as ILokiLabelFilterNode)
                                )
                            }}
                        />
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'row',
                            alignItems: 'center',
                            gap: 10,
                        }}
                    >
                        <Label style={{ flexShrink: 0 }}>{t('max lines')}</Label>
                        <Input
                            size='mini'
                            overrides={{
                                Root: {
                                    style: {
                                        width: 80,
                                    },
                                },
                            }}
                            value={tempMaxLines}
                            onChange={(e) => {
                                const v = (e.target as HTMLInputElement).value ?? '0'
                                setTempMaxLines(parseInt(v, 10))
                            }}
                            onKeyPress={(e) => {
                                if (e.key === 'Enter') {
                                    setMaxLines(tempMaxLines)
                                }
                            }}
                        />
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'row',
                            alignItems: 'center',
                            flexShrink: 0,
                            gap: 10,
                        }}
                    >
                        <Label
                            style={{
                                flexShrink: 0,
                            }}
                        >
                            {t('search')}
                        </Label>
                        <Input
                            size='mini'
                            endEnhancer={<Search size='18px' />}
                            placeholder={t('please enter keywords')}
                            value={keyword_}
                            onChange={(e) => setKeyword_((e.target as HTMLInputElement).value ?? '')}
                            onKeyDown={(e) => {
                                if (e.key === 'Enter') {
                                    setKeyword(keyword_)
                                }
                            }}
                        />
                    </div>
                </div>
            }
        >
            <div
                style={{
                    width: '100%',
                    height: '100%',
                    overflow: 'hidden',
                    position: 'relative',
                }}
            >
                <GrafanaIFrame
                    className={styles.iframe}
                    style={style}
                    title='Loki Log'
                    baseUrl=''
                    pathname={pathname}
                    query={query}
                />
                <div
                    style={{
                        // display: showPlaster ? 'block' : 'none',
                        display: 'none',
                    }}
                >
                    <div
                        style={{
                            width: 306,
                            height: 69,
                            background: '#f7f8fa',
                            position: 'absolute',
                            top: 0,
                            left: 0,
                        }}
                    />
                    <div
                        style={{
                            height: 106,
                            background: '#f7f8fa',
                            position: 'absolute',
                            top: 67,
                            left: 5,
                            right: 299,
                        }}
                    />
                    <div
                        style={{
                            width: 10,
                            height: 100,
                            background: '#fff',
                            position: 'absolute',
                            top: 69,
                            right: 295,
                            border: '1px solid #dce1e6',
                            borderRadius: 3,
                            borderRightColor: '#fff',
                        }}
                    />
                </div>
            </div>
            <Modal
                overrides={{
                    Dialog: {
                        style: {
                            width: '80vw',
                            height: '80vh',
                            display: 'flex',
                            flexDirection: 'column',
                        },
                    },
                }}
                isOpen={openLineFiltersModal}
                onClose={() => setOpenLineFiltersModal(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('advanced search')}</ModalHeader>
                <ModalBody>
                    <LokiFiltersForm filters={lineFilterNodes} onSubmit={onLineFilterNodesSubmit} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
