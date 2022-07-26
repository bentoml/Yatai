import React from 'react'
import { BrowserRouter, Switch, Route } from 'react-router-dom'
import Header from '@/components/Header'
import OrganizationLayout from '@/components/OrganizationLayout'
import ClusterOverview from '@/pages/Cluster/Overview'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'
import Login from '@/pages/Yatai/Login'
import OrganizationApiTokens from '@/pages/Organization/ApiTokens'
import OrganizationClusters from '@/pages/Organization/Clusters'
import OrganizationMembers from '@/pages/Organization/Members'
import OrganizationDeployments from '@/pages/Organization/Deployments'
import OrganizationDeploymentForm from '@/pages/Organization/DeploymentForm'
import OrganizationSettings from '@/pages/Organization/Settings'
import ClusterDeployments from '@/pages/Cluster/Deployments'
import ClusterMembers from '@/pages/Cluster/Members'
import ClusterSettings from '@/pages/Cluster/Settings'
import ClusterLayout from '@/components/ClusterLayout'
import OrganizationBentoRepositories from '@/pages/Organization/BentoRepositories'
import OrganizationBentos from '@/pages/Organization/Bentos'
import OrganizationModels from '@/pages/Organization/Models'
import OrganizationEvents from '@/pages/Organization/Events'
import BentoRepositoryOverview from '@/pages/BentoRepository/Overview'
import BentoRepositoryBentos from '@/pages/BentoRepository/Bentos'
import DeploymentOverview from '@/pages/Deployment/Overview'
import DeploymentRevisions from '@/pages/Deployment/Revisions'
import DeploymentTerminalRecordPlayer from '@/pages/Deployment/TerminalRecordPlayer'
import DeploymentReplicas from '@/pages/Deployment/Replicas'
import DeploymentLog from '@/pages/Deployment/Log'
import DeploymentMonitor from '@/pages/Deployment/Monitor'
import DeploymentEdit from '@/pages/Deployment/Edit'
import DeploymentRevisionRollback from '@/pages/Deployment/RevisionRollback'
import BentoRepositoryLayout from '@/components/BentoRepositoryLayout'
import DeploymentLayout from '@/components/DeploymentLayout'
import ModelRepositoryLayout from '@/components/ModelRepositoryLayout'
import ModelRepositoryOverview from '@/pages/ModelRepository/Overview'
import ModelRepositoryModels from '@/pages/ModelRepository/Models'
import OrganizationModelRepositories from '@/pages/Organization/ModelRepositories'
import { ChatWidget } from '@papercups-io/chat-widget'
import ModelLayout from '@/components/ModelLayout'
import ModelOverview from '@/pages/Model/Overview'
import BentoLayout from '@/components/BentoLayout'
import BentoOverview from '@/pages/Bento/Overview'
import BentoRepositoryDeployments from '@/pages/BentoRepository/Deployments'
import Home from '@/pages/Yatai/Home'
import Setup from '@/pages/Yatai/Setup'

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
                    <Route exact path='/bento_repositories/:bentoRepositoryName/bentos/:bentoVersion/:path?/:path?'>
                        <BentoLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/bento_repositories/:bentoRepositoryName/bentos/:bentoVersion'
                                    component={BentoOverview}
                                />
                            </Switch>
                        </BentoLayout>
                    </Route>
                    <Route exact path='/bento_repositories/:bentoRepositoryName/:path?/:path?'>
                        <BentoRepositoryLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/bento_repositories/:bentoRepositoryName'
                                    component={BentoRepositoryOverview}
                                />
                                <Route
                                    exact
                                    path='/bento_repositories/:bentoRepositoryName/bentos'
                                    component={BentoRepositoryBentos}
                                />
                                <Route
                                    exact
                                    path='/bento_repositories/:bentoRepositoryName/deployments'
                                    component={BentoRepositoryDeployments}
                                />
                            </Switch>
                        </BentoRepositoryLayout>
                    </Route>
                    <Route
                        exact
                        path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/:path?/edit'
                    >
                        <OrganizationLayout>
                            <Route
                                exact
                                path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/edit'
                                component={DeploymentEdit}
                            />
                        </OrganizationLayout>
                    </Route>
                    <Route
                        exact
                        path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/:path?/revisions/:path?/rollback'
                    >
                        <OrganizationLayout>
                            <Route
                                exact
                                path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/revisions/:revisionUid/rollback'
                                component={DeploymentRevisionRollback}
                            />
                        </OrganizationLayout>
                    </Route>
                    <Route
                        exact
                        path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/:path?/:path?'
                    >
                        <DeploymentLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName'
                                    component={DeploymentOverview}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/revisions'
                                    component={DeploymentRevisions}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/replicas'
                                    component={DeploymentReplicas}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/log'
                                    component={DeploymentLog}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/monitor'
                                    component={DeploymentMonitor}
                                />
                                <Route
                                    exact
                                    path='/clusters/:clusterName/namespaces/:kubeNamespace/deployments/:deploymentName/terminal_records/:uid'
                                    component={DeploymentTerminalRecordPlayer}
                                />
                            </Switch>
                        </DeploymentLayout>
                    </Route>
                    <Route exact path='/clusters/:clusterName/:path?/:path?'>
                        <ClusterLayout>
                            <Switch>
                                <Route exact path='/clusters/:clusterName' component={ClusterOverview} />
                                <Route exact path='/clusters/:clusterName/deployments' component={ClusterDeployments} />
                                <Route exact path='/clusters/:clusterName/members' component={ClusterMembers} />
                                <Route exact path='/clusters/:clusterName/settings' component={ClusterSettings} />
                            </Switch>
                        </ClusterLayout>
                    </Route>
                    <Route exact path='/model_repositories/:modelRepositoryName/models/:modelVersion/:path?/:path?'>
                        <ModelLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/model_repositories/:modelRepositoryName/models/:modelVersion'
                                    component={ModelOverview}
                                />
                            </Switch>
                        </ModelLayout>
                    </Route>
                    <Route exact path='/model_repositories/:modelRepositoryName/:path?/:path?'>
                        <ModelRepositoryLayout>
                            <Switch>
                                <Route
                                    exact
                                    path='/model_repositories/:modelRepositoryName'
                                    component={ModelRepositoryOverview}
                                />
                                <Route
                                    exact
                                    path='/model_repositories/:modelRepositoryName/models'
                                    component={ModelRepositoryModels}
                                />
                            </Switch>
                        </ModelRepositoryLayout>
                    </Route>
                    <Route exact path='/login' component={Login} />
                    <Route exact path='/setup' component={Setup} />
                    <Route>
                        <OrganizationLayout>
                            <Switch>
                                <Route exact path='/' component={Home} />
                                <Route exact path='/bentos' component={OrganizationBentos} />
                                <Route exact path='/models' component={OrganizationModels} />
                                <Route exact path='/events' component={OrganizationEvents} />
                                <Route exact path='/api_tokens' component={OrganizationApiTokens} />
                                <Route exact path='/clusters' component={OrganizationClusters} />
                                <Route exact path='/members' component={OrganizationMembers} />
                                <Route exact path='/bento_repositories' component={OrganizationBentoRepositories} />
                                <Route exact path='/model_repositories' component={OrganizationModelRepositories} />
                                <Route exact path='/deployments' component={OrganizationDeployments} />
                                <Route exact path='/new_deployment' component={OrganizationDeploymentForm} />
                                <Route exact path='/settings' component={OrganizationSettings} />
                            </Switch>
                        </OrganizationLayout>
                    </Route>
                </Switch>
                <ChatWidget
                    token='25ad5fd9-293b-4e0f-9601-5b0cd7846b48'
                    inbox='ac3ebd50-fc10-4299-9a1c-496841b49a6f'
                    title='Welcome to Yatai👋 👋 👋'
                    subtitle='Ask us questions or give us feedback - we will reply ASAP!😊'
                    primaryColor='#47AFD1'
                    newMessagePlaceholder='Start typing...'
                    showAgentAvailability={false}
                    agentAvailableText='We are online right now!'
                    agentUnavailableText='We are away at the moment.'
                    requireEmailUpfront={false}
                    iconVariant='outlined'
                    baseUrl='https://yatai-community-papercups.herokuapp.com'
                    // Optionally include data about your customer here to identify them
                    // customer={{
                    //   name: __CUSTOMER__.name,
                    //   email: __CUSTOMER__.email,
                    //   external_id: __CUSTOMER__.id,
                    //   metadata: {
                    //     plan: "premium"
                    //   }
                    // }}
                />
            </div>
        </BrowserRouter>
    )
}

export default Routes
