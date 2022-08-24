/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable no-case-declarations */
import { fetchNews } from '@/services/news'
import React, { useEffect, useMemo, useState } from 'react'
import { createUseStyles } from 'react-jss'
import { useQuery } from 'react-query'
import { Notification } from 'baseui/notification'
import { INewsItem } from '@/schemas/news'
import { StyledLink } from 'baseui/link'
import useTranslation from '@/hooks/useTranslation'
import Card from '@/components/Card'
import { FiActivity } from 'react-icons/fi'
import List from '@/components/List'
import { ListItem } from 'baseui/list'
import { listOrganizationEvents } from '@/services/organization'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import Time from '@/components/Time'
import { listOrganizationDeployments } from '@/services/deployment'
import { IDeploymentSchema } from '@/schemas/deployment'
import { useStyletron } from 'baseui'
import DeploymentStatusTag from '@/components/DeploymentStatusTag'
import { GrBlog, GrDocumentText } from 'react-icons/gr'
import { BiNotepad } from 'react-icons/bi'
import Image from 'rc-image'
import EventList from '@/components/EventList'

const useStyles = createUseStyles({
    root: {},
    notification: {
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
    },
})

const makeNotificationReadedKey = (n: INewsItem) => `notification_readed:${n.title}`

export default function Home() {
    const styles = useStyles()
    const newsInfo = useQuery('news', fetchNews)
    const [notification, setNotification] = useState<INewsItem>()
    const eventsInfo = useQuery('recent_events', () => listOrganizationEvents({ start: 0, count: 4 }))
    const deploymentsInfo = useQuery('recent_deployments', () =>
        listOrganizationDeployments({ start: 0, count: 4, q: 'sort:updated_at-desc' })
    )
    const [t] = useTranslation()
    const [, theme] = useStyletron()

    useEffect(() => {
        let timer: number
        let idx = 0
        const tick = () => {
            const notifications = (newsInfo.data?.notifications ?? []).filter(
                (n) => !window.localStorage.getItem(makeNotificationReadedKey(n))
            )
            if (idx >= notifications.length) {
                idx = 0
            }
            const n = notifications[idx]
            setNotification(n)
            idx++
            timer = window.setTimeout(tick, 5000)
        }
        tick()
        return () => {
            window.clearTimeout(timer)
        }
    }, [newsInfo.data?.notifications])

    const listItemOverrides = useMemo(
        () => ({
            Content: {
                style: {
                    minHeight: '48px',
                },
            },
        }),
        []
    )

    return (
        <div className={styles.root}>
            {notification && (
                <div className={styles.notification}>
                    <Notification
                        overrides={{
                            Body: {
                                style: {
                                    width: '100%',
                                    maxWidth: '500px',
                                    margin: '0px 0px 20px 0px',
                                },
                            },
                        }}
                        closeable
                        onClose={() => {
                            window.localStorage.setItem(makeNotificationReadedKey(notification), 'true')
                            setNotification(undefined)
                        }}
                        kind={notification.level || 'info'}
                    >
                        {notification.link ? (
                            <StyledLink href={notification.link} target='_blank'>
                                {notification.title}
                            </StyledLink>
                        ) : (
                            notification.title
                        )}
                    </Notification>
                </div>
            )}
            <Card title={t('recent activities')} titleIcon={FiActivity}>
                <EventList events={eventsInfo.data?.items ?? []} isLoading={eventsInfo.isLoading} />
            </Card>
            <Card title={t('active deployments')} titleIcon={resourceIconMapping.deployment}>
                <List
                    isLoading={deploymentsInfo.isLoading}
                    items={deploymentsInfo.data?.items ?? []}
                    emptyText={t('no active deployments')}
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
                                        <Link
                                            to={`/clusters/${item.cluster?.name}/namespaces/${item.kube_namespace}/deployments/${item.name}`}
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
            <div
                style={{
                    display: 'grid',
                    gap: 20,
                    gridTemplateColumns: '1fr 1fr 1fr',
                }}
            >
                <Card title={t('documentation')} titleIcon={GrDocumentText}>
                    <List
                        isLoading={newsInfo.isLoading}
                        items={newsInfo.data?.documentations ?? []}
                        onRenderItem={(item: INewsItem) => {
                            const artwork = () => (item.cover ? <Image src={item.cover} height={32} /> : '◼︎')
                            return (
                                <ListItem overrides={listItemOverrides} sublist artwork={artwork}>
                                    <div>
                                        <StyledLink href={item.link} target='_blank'>
                                            {item.title}
                                        </StyledLink>
                                    </div>
                                </ListItem>
                            )
                        }}
                    />
                </Card>
                <Card title={t('release notes')} titleIcon={BiNotepad}>
                    <List
                        isLoading={newsInfo.isLoading}
                        items={newsInfo.data?.release_notes ?? []}
                        onRenderItem={(item: INewsItem) => {
                            const artwork = () => (item.cover ? <Image src={item.cover} height={32} /> : '◼︎')
                            return (
                                <ListItem overrides={listItemOverrides} sublist artwork={artwork}>
                                    <div>
                                        <StyledLink href={item.link} target='_blank'>
                                            {item.title}
                                        </StyledLink>
                                    </div>
                                </ListItem>
                            )
                        }}
                    />
                </Card>
                <Card title={t('blog posts')} titleIcon={GrBlog}>
                    <List
                        isLoading={newsInfo.isLoading}
                        items={newsInfo.data?.blog_posts ?? []}
                        onRenderItem={(item: INewsItem) => {
                            const artwork = () => (item.cover ? <Image src={item.cover} height={32} /> : '◼︎')
                            return (
                                <ListItem overrides={listItemOverrides} sublist artwork={artwork}>
                                    <div>
                                        <StyledLink href={item.link} target='_blank'>
                                            {item.title}
                                        </StyledLink>
                                    </div>
                                </ListItem>
                            )
                        }}
                    />
                </Card>
            </div>
        </div>
    )
}
