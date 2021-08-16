import React, { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import { fetchCurrentUser } from '@/services/user'
import { useQuery } from 'react-query'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import axios from 'axios'
import { toaster } from 'baseui/toast'
import { getErrMsg } from '@/utils/error'
import qs from 'qs'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Link, useHistory, useLocation } from 'react-router-dom'
import { useStyletron } from 'baseui'
import { headerHeight, resourceIconMapping } from '@/consts'
import { SidebarContext } from '@/contexts/SidebarContext'
import { StyledSpinnerNext } from 'baseui/spinner'
import logo from '@/assets/logo.png'
import useTranslation from '@/hooks/useTranslation'
import { createOrganization } from '@/services/organization'
import { Select, SIZE as SelectSize } from 'baseui/select'
import { useOrganization } from '@/hooks/useOrganization'
import OrganizationForm from '@/components/OrganizationForm'
import { ICreateOrganizationSchema } from '@/schemas/organization'
import { BiMoon, BiSun } from 'react-icons/bi'
import color from 'color'
import { createUseStyles } from 'react-jss'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { useThemeType } from '@/hooks/useThemeType'
import classNames from 'classnames'
import User from '@/components/User'
import Text from '@/components/Text'
import { ICreateClusterSchema } from '@/schemas/cluster'
import { createCluster } from '@/services/cluster'
import { useCluster } from '@/hooks/useCluster'
import ClusterForm from '@/components/ClusterForm'
import { useFetchClusters } from '@/hooks/useFetchClusters'
import { useFetchOrganizations } from '@/hooks/useFetchOrganizations'

const useThemeToggleStyles = createUseStyles({
    root: ({ theme }: IThemedStyleProps) => ({
        position: 'relative',
        cursor: 'pointer',
        border: `1px solid ${theme.borders.border300.borderColor}`,
        borderRadius: 18,
        height: 18,
    }),
    track: () => ({
        height: 18,
        padding: '0 4px',
        transition: 'all 0.2s ease',
        display: 'flex',
        flexDirection: 'row',
        alignItems: 'center',
    }),
    thumb: ({ theme }: IThemedStyleProps) => ({
        position: 'absolute',
        height: 18,
        width: 18,
        padding: 1,
        top: -1,
        left: -2,
        borderRadius: '50%',
        background: theme.colors.contentPrimary,
        color: theme.colors.backgroundPrimary,
        transition: 'all 0.5s cubic-bezier(0.23, 1, 0.32, 1) 0ms',
        transform: 'translateX(0)',
        textAlign: 'center',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
    }),
    checked: () => ({
        transform: 'translateX(24px)',
    }),
})

interface IThemeToggleProps {
    className?: string
}

const ThemeToggle = ({ className }: IThemeToggleProps) => {
    const [, theme] = useStyletron()
    const themeType = useCurrentThemeType()
    const styles = useThemeToggleStyles({ theme, themeType })
    const { setThemeType } = useThemeType()
    const checked = themeType === 'dark'

    return (
        <div
            role='button'
            tabIndex={0}
            className={classNames(className, styles.root)}
            onClick={() => {
                const newThemeType = themeType === 'dark' ? 'light' : 'dark'
                setThemeType(newThemeType)
            }}
        >
            <div className={styles.track}>
                <BiSun />
                <BiMoon style={{ marginLeft: 4 }} />
            </div>
            <div className={classNames({ [styles.thumb]: true, [styles.checked]: checked })}>
                {!checked ? <BiSun /> : <BiMoon />}
            </div>
        </div>
    )
}

const orgPathPattern = /(\/orgs\/)([^/]+)(.*)/
const clusterPathPattern = /\/orgs\/[^/]+\/clusters\/([^/]+).*/

export default function Header() {
    const location = useLocation()
    const history = useHistory()
    // FIXME: can not use useParams, because of Header is not under the Route component
    const orgMatch = useMemo(() => location.pathname.match(orgPathPattern), [location.pathname])
    const clusterMatch = useMemo(() => location.pathname.match(clusterPathPattern), [location.pathname])
    const orgName = orgMatch ? orgMatch[2] : undefined
    const clusterName = clusterMatch ? clusterMatch[1] : undefined

    const errMsgExpireTimeSeconds = 5
    const lastErrMsgRef = useRef<Record<string, number>>({})
    const lastLocationPathRef = useRef(location.pathname)

    useEffect(() => {
        if (lastLocationPathRef.current !== location.pathname) {
            lastErrMsgRef.current = {}
        }
        lastLocationPathRef.current = location.pathname
    }, [location.pathname])

    useEffect(() => {
        // eslint-disable-next-line @typescript-eslint/no-explicit-any
        if ((axios.interceptors.response as any).handlers.length > 0) {
            return
        }
        axios.interceptors.response.use(
            (response) => {
                return response
            },
            (error) => {
                const errMsg = getErrMsg(error)
                if (error.response?.status === 403 && error.config.method === 'get') {
                    const search = qs.parse(location.search, { ignoreQueryPrefix: true })
                    let { redirect } = search
                    if (redirect && typeof redirect === 'string') {
                        redirect = decodeURI(redirect)
                    } else if (['/login', '/logout'].indexOf(location.pathname) < 0) {
                        redirect = `${location.pathname}${location.search}`
                    } else {
                        redirect = '/'
                    }
                    window.location.href = `${window.location.protocol}//${
                        window.location.host
                    }/oauth/github?redirect=${encodeURIComponent(redirect)}`
                } else if (Date.now() - (lastErrMsgRef.current[errMsg] || 0) > errMsgExpireTimeSeconds * 1000) {
                    toaster.negative(errMsg, { autoHideDuration: (errMsgExpireTimeSeconds + 1) * 1000 })
                    lastErrMsgRef.current[errMsg] = Date.now()
                }
                return Promise.reject(error)
            }
        )
        fetchCurrentUser()
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    const { currentUser, setCurrentUser } = useCurrentUser()
    const userInfo = useQuery('currentUser', fetchCurrentUser)
    useEffect(() => {
        if (userInfo.isSuccess) {
            setCurrentUser(userInfo.data)
        }
    }, [userInfo.data, userInfo.isSuccess, setCurrentUser])

    const { organization, setOrganization } = useOrganization()
    useEffect(() => {
        if (!orgName) {
            setOrganization(undefined)
        }
    }, [orgName, setOrganization])

    const orgsInfo = useFetchOrganizations({
        start: 0,
        count: 100,
    })

    const clustersInfo = useFetchClusters(orgName, { start: 0, count: 1000 })

    const { cluster, setCluster } = useCluster()
    useEffect(() => {
        if (!clusterName) {
            setCluster(undefined)
        }
    }, [clusterName, setCluster])

    const [css, theme] = useStyletron()
    const ctx = useContext(SidebarContext)
    const [t] = useTranslation()
    const generateLinkStyle = (path: string): React.CSSProperties => {
        const baseStyle = {
            textDecoration: 'none',
            marginBottom: -9,
            padding: '0px 6px 6px 6px',
            color: theme.colors.contentPrimary,
            borderBottom: '3px solid transparent',
            display: 'flex',
            alignItems: 'center',
            gap: 6,
        }
        if (location.pathname.startsWith(path)) {
            return {
                ...baseStyle,
                borderBottomColor: theme.colors.contentPrimary,
            }
        }
        return {
            ...baseStyle,
            borderBottomColor: 'transparent',
        }
    }

    const [isCreateOrgModalOpen, setIsCreateOrgModalOpen] = useState(false)

    const handleCreateOrg = useCallback(
        async (data: ICreateOrganizationSchema) => {
            await createOrganization(data)
            await orgsInfo.refetch()
            setIsCreateOrgModalOpen(false)
        },
        [orgsInfo]
    )

    const [isCreateClusterModalOpen, setIsCreateClusterModalOpen] = useState(false)

    const handleCreateCluster = useCallback(
        async (data: ICreateClusterSchema) => {
            if (!orgName) {
                return
            }
            await createCluster(orgName, data)
            await orgsInfo.refetch()
            setIsCreateClusterModalOpen(false)
        },
        [orgName, orgsInfo]
    )

    return (
        <header
            className={css({
                padding: '0 30px',
                position: 'fixed',
                background: color(theme.colors.backgroundPrimary).fade(0.5).rgb().string(),
                borderBottom: `1px solid ${theme.borders.border300.borderColor}`,
                backdropFilter: 'blur(10px)',
                zIndex: 1000,
                top: 0,
                height: `${headerHeight}px`,
                width: '100%',
                display: 'flex',
                flexFlow: 'row nowrap',
                boxSizing: 'border-box',
                alignItems: 'center',
            })}
        >
            <Link
                style={{
                    flex: '0 0 auto',
                    display: 'flex',
                    flexDirection: 'row',
                    textDecoration: 'none',
                    alignItems: 'center',
                    justifyContent: 'start',
                    boxSizing: 'border-box',
                    transition: 'width 200ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                }}
                to='/'
            >
                <div
                    style={{
                        flexShrink: 0,
                        display: 'flex',
                        justifyContent: 'center',
                        marginRight: 10,
                    }}
                >
                    <img
                        style={{
                            width: 38,
                            height: 36,
                            display: 'inline-flex',
                            transition: 'all 250ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                        }}
                        src={logo}
                        alt='logo'
                    />
                </div>
                {ctx.expanded && (
                    <Text
                        style={{
                            display: 'flex',
                            fontSize: '16px',
                            fontFamily: 'Zen Tokyo Zoo',
                        }}
                    >
                        YATAI
                    </Text>
                )}
            </Link>
            <div
                style={{
                    flexBasis: 1,
                    height: 20,
                    background: theme.colors.borderAlt,
                    margin: '0 20px',
                }}
            />
            <div
                style={{
                    flexShrink: 0,
                    display: 'flex',
                    gap: 10,
                    alignItems: 'center',
                }}
            >
                <Link
                    style={{
                        display: 'flex',
                        flexShrink: 0,
                        textDecoration: 'none',
                        gap: 8,
                        alignItems: 'center',
                        fontSize: '12px',
                    }}
                    to={orgName ? `/orgs/${orgName}` : '/'}
                >
                    {React.createElement(resourceIconMapping.organization, { size: 12 })}
                    <Text>{t('organization')}</Text>
                </Link>
                <div
                    style={{
                        width: 140,
                        flexShrink: 0,
                    }}
                >
                    <Select
                        isLoading={orgsInfo.isLoading}
                        clearable={false}
                        searchable={false}
                        options={
                            orgsInfo.data?.items.map((item) => ({
                                id: item.uid,
                                label: item.name,
                            })) ?? []
                        }
                        size={SelectSize.mini}
                        placeholder={t('select sth', [t('organization')])}
                        value={
                            organization && [
                                {
                                    id: organization.uid,
                                    label: organization.name,
                                },
                            ]
                        }
                        onChange={(v) => {
                            const org = orgsInfo.data?.items.find((item) => item.uid === v.option?.id)
                            if (org) {
                                history.push(`/orgs/${org.name}`)
                            }
                        }}
                    />
                </div>
                <Button
                    overrides={{
                        Root: {
                            style: {
                                flexShrink: 0,
                            },
                        },
                    }}
                    size={ButtonSize.mini}
                    onClick={() => {
                        setIsCreateOrgModalOpen(true)
                    }}
                >
                    {t('create')}
                </Button>
            </div>
            {cluster && (
                <>
                    <div
                        style={{
                            flexBasis: 1,
                            height: 20,
                            background: theme.colors.borderAlt,
                            margin: '0 20px',
                        }}
                    />
                    <Link
                        style={{
                            display: 'flex',
                            flexShrink: 0,
                            textDecoration: 'none',
                            gap: 8,
                            alignItems: 'center',
                            fontSize: '12px',
                        }}
                        to={orgName ? `/orgs/${orgName}/clusters/${cluster.name}` : '/'}
                    >
                        <div
                            style={{
                                display: 'flex',
                                flexShrink: 0,
                                gap: 8,
                                alignItems: 'center',
                                fontSize: '12px',
                            }}
                        >
                            {React.createElement(resourceIconMapping.cluster, { size: 12 })}
                            <Text>{t('cluster')}</Text>
                        </div>
                        <div
                            style={{
                                flexShrink: 0,
                                width: 140,
                            }}
                        >
                            <Select
                                isLoading={clustersInfo.isLoading}
                                clearable={false}
                                searchable={false}
                                options={
                                    clustersInfo.data?.items.map((item) => ({
                                        id: item.uid,
                                        label: item.name,
                                    })) ?? []
                                }
                                size={SelectSize.mini}
                                placeholder={t('select sth', [t('cluster')])}
                                value={
                                    cluster && [
                                        {
                                            id: cluster.uid,
                                            label: cluster.name,
                                        },
                                    ]
                                }
                                onChange={(v) => {
                                    const cluster_ = clustersInfo.data?.items.find((item) => item.uid === v.option?.id)
                                    if (cluster_) {
                                        history.push(`/orgs/${orgName}/clusters/${cluster_.name}`)
                                    }
                                }}
                            />
                        </div>
                        <Button
                            overrides={{
                                Root: {
                                    style: {
                                        flexShrink: 0,
                                        display: 'none',
                                    },
                                },
                            }}
                            size={ButtonSize.mini}
                            onClick={() => {
                                setIsCreateClusterModalOpen(true)
                            }}
                        >
                            {t('create')}
                        </Button>
                    </Link>
                </>
            )}
            <div style={{ flexGrow: 1 }} />
            <div
                className={css({
                    'flexShrink': 0,
                    'height': '100%',
                    'font-size': '14px',
                    'color': theme.colors.contentPrimary,
                    'display': 'flex',
                    'align-items': 'center',
                    'marginRight': '40px',
                    'gap': '30px',
                })}
            >
                {organization && (
                    <Link
                        style={generateLinkStyle(`/orgs/${organization.name}/bundles`)}
                        to={`/orgs/${organization.name}/bundles`}
                    >
                        {React.createElement(resourceIconMapping.bundle, { size: 12 })}
                        {t('bundle')}
                    </Link>
                )}
                {cluster && (
                    <Link
                        style={generateLinkStyle(`/orgs/${orgName}/clusters/${cluster.name}/deployments`)}
                        to={`/orgs/${orgName}/clusters/${cluster.name}/deployments`}
                    >
                        {React.createElement(resourceIconMapping.deployment, { size: 12 })}
                        {t('deployment')}
                    </Link>
                )}
                <ThemeToggle />
            </div>
            {currentUser ? <User user={currentUser} /> : <StyledSpinnerNext />}
            <Modal
                isOpen={isCreateOrgModalOpen}
                onClose={() => setIsCreateOrgModalOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('organization')])}</ModalHeader>
                <ModalBody>
                    <OrganizationForm onSubmit={handleCreateOrg} />
                </ModalBody>
            </Modal>
            <Modal
                isOpen={isCreateClusterModalOpen}
                onClose={() => setIsCreateClusterModalOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('create sth', [t('cluster')])}</ModalHeader>
                <ModalBody>
                    <ClusterForm onSubmit={handleCreateCluster} />
                </ModalBody>
            </Modal>
        </header>
    )
}
