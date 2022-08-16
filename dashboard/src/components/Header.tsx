import React, { useCallback, useContext, useEffect, useMemo, useRef, useState } from 'react'
import { changePassword, fetchCurrentUser } from '@/services/user'
import { useQuery } from 'react-query'
import { useCurrentUser } from '@/hooks/useCurrentUser'
import axios from 'axios'
import { toaster } from 'baseui/toast'
import { getErrMsg } from '@/utils/error'
import qs from 'qs'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import { Link, useLocation } from 'react-router-dom'
import { useStyletron } from 'baseui'
import { headerHeight, resourceIconMapping, yataiOrgHeader } from '@/consts'
import { SidebarContext } from '@/contexts/SidebarContext'
import logo from '@/assets/logo.svg'
import logoDark from '@/assets/logo-dark.svg'
import useTranslation from '@/hooks/useTranslation'
import { createOrganization, fetchCurrentOrganization } from '@/services/organization'
import { Select } from 'baseui/select'
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
import { IChangePasswordSchema } from '@/schemas/user'
import { createCluster } from '@/services/cluster'
import { useCluster } from '@/hooks/useCluster'
import ClusterForm from '@/components/ClusterForm'
import ReactCountryFlag from 'react-country-flag'
import i18n from '@/i18n'
import { simulationJump } from '@/utils'
import { FiLogOut } from 'react-icons/fi'
import { MdPassword } from 'react-icons/md'
import { useFetchOrganizations } from '@/hooks/useFetchOrganizations'
import { useFetchInfo } from '@/hooks/useFetchInfo'
import { Button } from 'baseui/button'
import PasswordForm from './PasswordForm'

const useStyles = createUseStyles({
    userWrapper: {
        'position': 'relative',
        'cursor': 'pointer',
        'display': 'flex',
        'align-items': 'center',
        'min-width': '140px',
        'height': '100%',
        'margin-left': '20px',
        'flex-direction': 'column',
        '&:hover': {
            '& $userMenu': {
                display: 'flex',
            },
        },
    },
    userAvatarWrapper: {
        'height': '100%',
        'display': 'flex',
        'align-items': 'center',
    },
    userMenu: (props: IThemedStyleProps) => ({
        'background': props.theme.colors.backgroundPrimary,
        'position': 'absolute',
        'top': '100%',
        'display': 'none',
        'margin': 0,
        'padding': 0,
        'line-height': 1.6,
        'flex-direction': 'column',
        'width': '100%',
        'font-size': '13px',
        'box-shadow': props.theme.lighting.shadow400,
        '& a': {
            '&:link': {
                'color': props.theme.colors.contentPrimary,
                'text-decoration': 'none',
            },
            '&:hover': {
                'color': props.theme.colors.contentPrimary,
                'text-decoration': 'none',
            },
            '&:visited': {
                'color': props.theme.colors.contentPrimary,
                'text-decoration': 'none',
            },
        },
    }),
    userMenuItem: (props: IThemedStyleProps) => ({
        'padding': '8px 12px',
        'display': 'flex',
        'align-items': 'center',
        'gap': '10px',
        '&:hover': {
            background: color(props.theme.colors.backgroundPrimary)
                .darken(props.themeType === 'light' ? 0.06 : 0.2)
                .rgb()
                .string(),
        },
    }),
})

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

const clusterPathPattern = /\/clusters\/([^/]+).*/

export default function Header() {
    const [css, theme] = useStyletron()
    const themeType = useCurrentThemeType()
    const styles = useStyles({ theme, themeType })
    const location = useLocation()
    // FIXME: can not use useParams, because of Header is not under the Route component
    const clusterMatch = useMemo(() => location.pathname.match(clusterPathPattern), [location.pathname])
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
                    } else if (['/login', '/logout', '/setup'].indexOf(location.pathname) < 0) {
                        redirect = `${location.pathname}${location.search}`
                    } else {
                        redirect = '/'
                    }
                    if (
                        location.pathname !== '/login' &&
                        location.pathname !== '/login/' &&
                        location.pathname !== '/setup' &&
                        location.pathname !== '/setup/'
                    ) {
                        window.location.href = `${window.location.protocol}//${
                            window.location.host
                        }/login?redirect=${encodeURIComponent(redirect)}`
                    }
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
        axios.defaults.headers.common[yataiOrgHeader] = organization?.name ?? ''
    }, [organization])

    const orgsInfo = useFetchOrganizations({
        start: 0,
        count: 100,
    })

    const currentOrgInfo = useQuery('currentOrg', fetchCurrentOrganization)

    useEffect(() => {
        if (!organization && currentOrgInfo.isSuccess && currentOrgInfo.data) {
            setOrganization(currentOrgInfo.data)
        }
    }, [currentOrgInfo.data, currentOrgInfo.isSuccess, organization, setOrganization])

    const infoInfo = useFetchInfo()

    const showOrgSelector = useMemo(() => {
        return currentUser?.is_super_admin && infoInfo.data?.is_sass && currentOrgInfo.data?.name === 'default'
    }, [currentOrgInfo.data?.name, currentUser?.is_super_admin, infoInfo.data?.is_sass])

    const { setCluster } = useCluster()
    useEffect(() => {
        if (!clusterName) {
            setCluster(undefined)
        }
    }, [clusterName, setCluster])

    const ctx = useContext(SidebarContext)
    const [t] = useTranslation()

    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const handleRenderLanguageOption = useCallback(({ option }: any) => {
        return (
            <div>
                {option.flag && <span style={{ marginRight: 8, verticalAlign: 'middle' }}>{option.flag}</span>}
                <span style={{ verticalAlign: 'middle' }}>{option.text}</span>
            </div>
        )
    }, [])

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

    const handleCreateCluster = useCallback(async (data: ICreateClusterSchema) => {
        await createCluster(data)
        setIsCreateClusterModalOpen(false)
    }, [])

    const [isChangePasswordOpen, setIsChangePasswordOpen] = useState(false)

    const handleChangePassword = useCallback(
        async (data: IChangePasswordSchema) => {
            await changePassword(data)
            setIsChangePasswordOpen(false)
            toaster.positive(t('password changed'), { autoHideDuration: 2000 })
        },
        [t]
    )

    const currentThemeType = useCurrentThemeType()

    return (
        <header
            className={css({
                padding: '0 23px',
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
                    gap: 12,
                }}
                to='/'
            >
                <div
                    style={{
                        flexShrink: 0,
                        display: 'flex',
                        justifyContent: 'center',
                    }}
                >
                    <img
                        style={{
                            width: 26,
                            height: 26,
                            display: 'inline-flex',
                            transition: 'all 250ms cubic-bezier(0.7, 0.1, 0.33, 1) 0ms',
                        }}
                        src={currentThemeType === 'light' ? logo : logoDark}
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
            {!showOrgSelector && organization && (
                <>
                    <div
                        style={{
                            flexBasis: 1,
                            flexShrink: 0,
                            height: 20,
                            background: theme.colors.borderOpaque,
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
                                gap: 6,
                                alignItems: 'center',
                            }}
                            to='/'
                        >
                            {React.createElement(resourceIconMapping.organization, { size: 12 })}
                            <Text
                                style={{
                                    fontFamily: 'Teko',
                                    fontSize: '18px',
                                }}
                            >
                                {organization?.name}
                            </Text>
                        </Link>
                    </div>
                </>
            )}
            {showOrgSelector && (
                <>
                    <div
                        style={{
                            flexBasis: 1,
                            flexShrink: 0,
                            height: 20,
                            background: theme.colors.borderOpaque,
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
                            to='/'
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
                                size='mini'
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
                                        setOrganization(org)
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
                            size='mini'
                            onClick={() => {
                                setIsCreateOrgModalOpen(true)
                            }}
                        >
                            {t('create')}
                        </Button>
                    </div>
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
                    'gap': '30px',
                })}
            >
                <ThemeToggle />
                <div
                    style={{
                        width: 140,
                    }}
                >
                    <Select
                        overrides={{
                            ControlContainer: {
                                style: {
                                    fontSize: 12,
                                },
                            },
                            InputContainer: {
                                style: {
                                    fontSize: 12,
                                },
                            },
                        }}
                        clearable={false}
                        searchable={false}
                        size='mini'
                        value={[{ id: i18n.language ? i18n.language.split('-')[0] : '' }]}
                        onChange={(params) => {
                            if (!params.option?.id) {
                                return
                            }
                            i18n.changeLanguage(params.option?.id as string)
                        }}
                        getOptionLabel={handleRenderLanguageOption}
                        getValueLabel={handleRenderLanguageOption}
                        options={[
                            {
                                id: 'en',
                                text: 'English',
                                flag: <ReactCountryFlag countryCode='US' svg />,
                            },
                            {
                                id: 'zh',
                                text: '中文',
                                flag: <ReactCountryFlag countryCode='CN' svg />,
                            },
                            {
                                id: 'ja',
                                text: '日本語',
                                flag: <ReactCountryFlag countryCode='JP' svg />,
                            },
                            {
                                id: 'ko',
                                text: '한국어',
                                flag: <ReactCountryFlag countryCode='KR' svg />,
                            },
                            {
                                id: 'vi',
                                text: 'Tiếng Việt',
                                flag: <ReactCountryFlag countryCode='VN' svg />,
                            },
                        ]}
                    />
                </div>
            </div>
            {currentUser && (
                <div className={styles.userWrapper}>
                    <div className={styles.userAvatarWrapper}>
                        <User user={currentUser} />
                    </div>
                    <div className={styles.userMenu}>
                        <Link className={styles.userMenuItem} to='/members'>
                            {React.createElement(resourceIconMapping.user_group, { size: 12 })}
                            <span>{t('members')}</span>
                        </Link>
                        <Link className={styles.userMenuItem} to='/api_tokens'>
                            {React.createElement(resourceIconMapping.api_token, { size: 12 })}
                            <span>{t('api tokens')}</span>
                        </Link>
                        <div
                            role='button'
                            tabIndex={0}
                            className={styles.userMenuItem}
                            onClick={() => {
                                setIsChangePasswordOpen(true)
                            }}
                        >
                            <MdPassword size={12} />
                            <span>{t('password')}</span>
                        </div>
                        <div
                            role='button'
                            tabIndex={0}
                            className={styles.userMenuItem}
                            onClick={() => {
                                simulationJump('/logout')
                            }}
                        >
                            <FiLogOut size={12} />
                            <span>{t('logout')}</span>
                        </div>
                    </div>
                </div>
            )}
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
            <Modal
                isOpen={isChangePasswordOpen}
                onClose={() => setIsChangePasswordOpen(false)}
                closeable
                animate
                autoFocus
            >
                <ModalHeader>{t('change password')}</ModalHeader>
                <ModalBody>
                    <PasswordForm onSubmit={handleChangePassword} />
                </ModalBody>
            </Modal>
        </header>
    )
}
