import React from 'react'
import { BrowserRouter, Switch, Route } from 'react-router-dom'
import Header from '@/components/Header'
import YataiLayout from '@/components/YataiLayout'
import OrganizationLayout from '@/components/OrganizationLayout'
import OrganizationOverview from '@/pages/Organization/Overview'
import ClusterOverview from '@/pages/Cluster/Overview'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import OrganizationClusters from '@/pages/Organization/Clusters'
import OrganizationMembers from '@/pages/Organization/Members'
import OrganizationDeployments from '@/pages/Organization/Deployments'
import ClusterYataiComponents from '@/pages/Cluster/YataiComponents'
import ClusterYataiComponentDetail from '@/pages/Cluster/YataiComponentDetail'
import ClusterDeployments from '@/pages/Cluster/Deployments'
import ClusterMembers from '@/pages/Cluster/Members'
import ClusterSettings from '@/pages/Cluster/Settings'
import ClusterLayout from '@/components/ClusterLayout'
import OrganizationBentos from '@/pages/Organization/Bentos'
import BentoOverview from '@/pages/Bento/Overview'
import BentoVersions from '@/pages/Bento/Versions'
import DeploymentOverview from '@/pages/Deployment/Overview'
import DeploymentSnapshots from '@/pages/Deployment/Snapshots'
import DeploymentTerminalRecordPlayer from '@/pages/Deployment/TerminalRecordPlayer'
import DeploymentLog from '@/pages/Deployment/Log'
import BentoLayout from '@/components/BentoLayout'
import UserProfile from '@/pages/Yatai/UserProfile'
import DeploymentLayout from '@/components/DeploymentLayout'
import ModelLayout from '@/components/ModelLayout'
import ModelOverview from '@/pages/Model/Overview'
import ModelVersions from '@/pages/Model/Versions'
import OrganizationModels from '@/pages/Organization/Models'

const useStyles = createUseStyles({
    'root': ({ theme }: IThemedStyleProps) => ({
        '& path': {
            stroke: theme.colors.contentPrimary,
        },
        ...Object.entries(theme.colors).reduce((p: Record<string, string>, [k, v]) => {
            return {
                ...p,
                [`--color-${k}`]: v,
            }
        }, {} as Record<string, string>),
    }),
    '@global': {
        '.react-lazylog': {
            background: 'var(--color-backgroundPrimary)',
        },
        '.react-lazylog-searchbar': {
            background: 'var(--color-backgroundPrimary)',
        },
        '.react-lazylog-searchbar-input': {
            background: 'var(--color-backgroundPrimary)',
        },
    },
})

const Routes = () => {
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const styles = useStyles({ theme, themeType })

    return (
        <BrowserRouter>
            <div
                className={styles.root}
                style={{
                    minHeight: '100vh',
                    background: themeType === 'light' ? '#fdfdfd' : theme.colors.backgroundSecondary,
                    color: theme.colors.contentPrimary,
                }}
            >
                <Header />
                <Switch>
                    <Route exact path='/bentos/:bentoName/:path?/:path?'>
                        <BentoLayout>
                            <Switch>
                                <Route exact path='/bentos/:bentoName' component={BentoOverview} />
                                <Route exact path='/bentos/:bentoName/versions' component={BentoVersions} />
                            </Switch>
                        </BentoLayout>
                    </Route>
                    <Route exact path='/clusters/:clusterName/deployments/:deploymentName/:path?/:path?'>
                        <DeploymentLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/clusters/:clusterName/deployments/:deploymentName'
                                    component={DeploymentOverview}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/deployments/:deploymentName/snapshots'
                                    component={DeploymentSnapshots}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/deployments/:deploymentName/log'
                                    component={DeploymentLog}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/deployments/:deploymentName/terminal_records/:uid'
                                    component={DeploymentTerminalRecordPlayer}
                                />
                            </Switch>
                        </DeploymentLayout>
                    </Route>
                    <Route exact path='/clusters/:clusterName/:path?/:path?'>
                        <ClusterLayout>
                            <Switch>
                                <Route exact path='/clusters/:clusterName' component={ClusterOverview} />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/yatai_components'
                                    component={ClusterYataiComponents}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/yatai_components/:componentType'
                                    component={ClusterYataiComponentDetail}
                                />
                                <Route exact path='/clusters/:clusterName/deployments' component={ClusterDeployments} />
                                <Route exact path='/clusters/:clusterName/members' component={ClusterMembers} />
                                <Route exact path='/clusters/:clusterName/settings' component={ClusterSettings} />
                            </Switch>
                        </ClusterLayout>
                    </Route>
                    <Route exact path='/models/:modelName/:path?/:path?'>
                        <ModelLayout>
                            <Switch>
                                <Route exact path='/models/:modelName' component={ModelOverview} />
                                <Route exact path='/models/:modelName/versions' component={ModelVersions} />
                            </Switch>
                        </ModelLayout>
                    </Route>
                    <Route>
                        <OrganizationLayout>
                            <Switch>
                                <Route exact path='/' component={OrganizationOverview} />
                                <Route exact path='/bentos' component={OrganizationBentos} />
                                <Route exact path='/clusters' component={OrganizationClusters} />
                                <Route exact path='/members' component={OrganizationMembers} />
                                <Route exact path='/models' component={OrganizationModels} />
                                <Route exact path='/deployments' component={OrganizationDeployments} />
                            </Switch>
                        </OrganizationLayout>
                        <YataiLayout>
                            <Switch>
                                <Route exact path='/user' component={UserProfile} />
                            </Switch>
                        </YataiLayout>
                    </Route>
                </Switch>
            </div>
        </BrowserRouter>
    )
}

export default Routes
