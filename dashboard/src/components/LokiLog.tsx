import { useDeployment } from '@/hooks/useDeployment'
import useTranslation, { Translator } from '@/hooks/useTranslation'
import { ILokiFilter } from '@/interfaces/ILoki'
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
import Card from './Card'
import GrafanaIFrame from './GrafanaIFrame'
import LokiFiltersForm from './LokiFiltersForm'
import Text from './Text'
import Label from './Label'

const filterToString = (filter: ILokiFilter): string => {
    const value = `"${_.replace(filter.value, /"/g, '\\"')}"`

    if (filter.type === 'contains') {
        if (filter.isRegexp) {
            return `|~ ${value}`
        }
        return `|= ${value}`
    }
    if (filter.isRegexp) {
        return `!~ ${value}`
    }
    return `!= ${value}`
}

const filtersToReadableString = (t: Translator, filters: ILokiFilter[]): string => {
    return filters
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

    const [keyword, setKeyword] = useState<string | undefined>()
    const [keyword_, setKeyword_] = useState<string | undefined>()

    const [filters, setFilters] = useState<ILokiFilter[]>([])
    const [openFiltersModal, setOpenFiltersModal] = useState(false)

    const onFiltersSubmit = useCallback((filters_) => {
        setFilters(filters_)
        setOpenFiltersModal(false)
    }, [])

    const [, theme] = useStyletron()

    const [tempMaxLines, setTempMaxLines] = useState(1000)
    const [maxLines, setMaxLines] = useState(tempMaxLines)

    const grafanaRootPath = cluster?.grafana_root_path
    const dataSource = 'Loki'

    if (!grafanaRootPath) {
        return <div>no data</div>
    }

    let labels: string[] = []
    let defaultFilters: ILokiFilter[] = []
    if (deployment) {
        labels = [`yatai_ai_deployment="${deployment.name}"`]
        defaultFilters = [
            {
                type: 'not contains',
                value: 'kube-probe',
                isRegexp: false,
            },
        ]
    }

    if (podName) {
        labels = [...labels, `pod="${podName}"`]
    }

    if (namespace) {
        labels = [...labels, `namespace="${namespace}"`]
    }

    if (keyword) {
        defaultFilters = [
            ...defaultFilters,
            {
                type: 'contains',
                value: `(?i)${keyword}`,
                isRegexp: true,
            },
        ]
    }

    const expr = `{${labels.join(', ')}} ${[...defaultFilters, ...filters].map(filterToString).join(' ')}`

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

    const filtersReadableString = filtersToReadableString(t, filters)

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
                                marginRight: 10,
                            }}
                        >
                            {filtersReadableString}
                        </Text>
                    </div>
                    <div
                        style={{
                            marginRight: 10,
                        }}
                    >
                        <Button size='mini' onClick={() => setOpenFiltersModal(true)}>
                            {t('advanced search')}
                        </Button>
                    </div>
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'row',
                            alignItems: 'center',
                            marginRight: 10,
                        }}
                    >
                        <Label style={{ marginRight: 10 }}>{t('max lines')}</Label>
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
                        }}
                    >
                        <Label style={{ marginRight: 10 }}>{t('search')}</Label>
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
                isOpen={openFiltersModal}
                onClose={() => setOpenFiltersModal(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('advanced search')}</ModalHeader>
                <ModalBody>
                    <LokiFiltersForm filters={filters} onSubmit={onFiltersSubmit} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
