import React from 'react'
import { BrowserRouter, Switch, Route } from 'react-router-dom'
import Header from '@/components/Header'
import YataiLayout from '@/components/YataiLayout'
import Home from '@/pages/Yatai/Home'
import OrganizationLayout from '@/components/OrganizationLayout'
import OrganizationOverview from '@/pages/Organization/Overview'
import ClusterOverview from '@/pages/Cluster/Overview'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import OrganizationClusters from '@/pages/Organization/Clusters'
import OrganizationMembers from '@/pages/Organization/Members'
import ClusterDeployments from '@/pages/Cluster/Deployments'
import ClusterMembers from '@/pages/Cluster/Members'
import ClusterLayout from '@/components/ClusterLayout'
import OrganizationBentos from '@/pages/Organization/Bentos'
import BentoOverview from '@/pages/Bento/Overview'
import BentoVersions from '@/pages/Bento/Versions'
import DeploymentOverview from '@/pages/Deployment/Overview'
import DeploymentSnapshots from '@/pages/Deployment/Snapshots'
import DeploymentTerminalRecordPlayer from '@/pages/Deployment/TerminalRecordPlayer'
import BentoLayout from '@/components/BentoLayout'
import UserProfile from '@/pages/Yatai/UserProfile'
import DeploymentLayout from '@/components/DeploymentLayout'

const useStyles = createUseStyles({
    root: ({ theme }: IThemedStyleProps) => ({
        '& path': {
            stroke: theme.colors.contentPrimary,
        },
    }),
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
                    <Route exact path='/orgs/:orgName/bentos/:bentoName/:path?/:path?'>
                        <BentoLayout>
                            <Switch>
                                <Route exact path='/orgs/:orgName/bentos/:bentoName' component={BentoOverview} />
                                <Route
                                    exact
                                    path='/orgs/:orgName/bentos/:bentoName/versions'
                                    component={BentoVersions}
                                />
                            </Switch>
                        </BentoLayout>
                    </Route>
                    <Route exact path='/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/:path?/:path?'>
                        <DeploymentLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName'
                                    component={DeploymentOverview}
                                />
                                <Route
                                    exact
                                    path='/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/snapshots'
                                    component={DeploymentSnapshots}
                                />
                                <Route
                                    exact
                                    path='/orgs/:orgName/clusters/:clusterName/deployments/:deploymentName/terminal_records/:uid'
                                    component={DeploymentTerminalRecordPlayer}
                                />
                            </Switch>
                        </DeploymentLayout>
                    </Route>
                    <Route exact path='/orgs/:orgName/clusters/:clusterName/:path?/:path?'>
                        <ClusterLayout>
                            <Switch>
                                <Route exact path='/orgs/:orgName/clusters/:clusterName' component={ClusterOverview} />
                                <Route
                                    exact
                                    path='/orgs/:orgName/clusters/:clusterName/deployments'
                                    component={ClusterDeployments}
                                />
                                <Route
                                    exact
                                    path='/orgs/:orgName/clusters/:clusterName/members'
                                    component={ClusterMembers}
                                />
                            </Switch>
                        </ClusterLayout>
                    </Route>
                    <Route exact path='/orgs/:orgName/:path?/:path?'>
                        <OrganizationLayout>
                            <Switch>
                                <Route exact path='/orgs/:orgName' component={OrganizationOverview} />
                                <Route exact path='/orgs/:orgName/bentos' component={OrganizationBentos} />
                                <Route exact path='/orgs/:orgName/clusters' component={OrganizationClusters} />
                                <Route exact path='/orgs/:orgName/members' component={OrganizationMembers} />
                            </Switch>
                        </OrganizationLayout>
                    </Route>
                    <Route>
                        <YataiLayout>
                            <Switch>
                                <Route exact path='/user' component={UserProfile} />
                                <Route exact path='/' component={Home} />
                            </Switch>
                        </YataiLayout>
                    </Route>
                </Switch>
            </div>
        </BrowserRouter>
    )
}

export default Routes
