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
                    background: theme.colors.backgroundPrimary,
                    color: theme.colors.contentPrimary,
                }}
            >
                <Header />
                <Switch>
                    <Route exact path='/orgs/:orgName/:path?/:path?'>
                        <OrganizationLayout>
                            <Switch>
                                <Route exact path='/orgs/:orgName' component={OrganizationOverview} />
                                <Route exact path='/orgs/:orgName/clusters' component={OrganizationClusters} />
                                <Route exact path='/orgs/:orgName/members' component={OrganizationMembers} />
                                <Route exact path='/orgs/:orgName/clusters/:clusterName/:path?/:path?'>
                                    <OrganizationLayout>
                                        <Switch>
                                            <Route
                                                exact
                                                path='/orgs/:orgName/clusters/:clusterName'
                                                component={ClusterOverview}
                                            />
                                        </Switch>
                                    </OrganizationLayout>
                                </Route>
                            </Switch>
                        </OrganizationLayout>
                    </Route>
                    <Route>
                        <YataiLayout>
                            <Switch>
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
