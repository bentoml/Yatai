import { useCluster } from '@/hooks/useCluster'
import { IDeploymentSchema } from '@/schemas/deployment'
import React from 'react'
import GrafanaIFrame from './GrafanaIFrame'

interface IDeploymentMonitorProps {
    deployment: IDeploymentSchema
}

export default function DeploymentMonitor({ deployment }: IDeploymentMonitorProps) {
    const iframeStyle = {
        height: 300,
    }

    const { cluster } = useCluster()

    if (!cluster) {
        return <div>no data</div>
    }

    const podsNetworkingGrafanaExternalPathname = `${cluster.grafana_root_path}d/networking-pod-2/kubernetes-networking-pod-2`
    const podsNetworkingGrafanaPathname = `${cluster.grafana_root_path}d-solo/networking-pod-2/kubernetes-networking-pod-2`
    const podsResourcesGrafanaExternalPathname = `${cluster.grafana_root_path}d/resources-pod-2/kubernetes-compute-resources-pod-2`
    const podsResourcesGrafanaPathname = `${cluster.grafana_root_path}d-solo/resources-pod-2/kubernetes-compute-resources-pod-2`
    const baseGrafanaQuery = {
        'orgId': 1,
        'from': 'now-12h',
        'to': 'now',
        'var-datasource': 'Prometheus',
        'var-namespace': deployment.kube_namespace,
        'var-pod': deployment.name,
        'var-container': 'All',
        'fullscreen': true,
        'refresh': '30s',
    }

    return (
        <div>
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'row',
                    justifyContent: 'center',
                    gap: 2,
                }}
            >
                <GrafanaIFrame
                    title='Current Rate of Bytes Received'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsNetworkingGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 3,
                    }}
                    externalPathname={podsNetworkingGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 3,
                    }}
                />
                <GrafanaIFrame
                    title='Current Rate of Bytes Transmitted'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsNetworkingGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 4,
                    }}
                    externalPathname={podsNetworkingGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 4,
                    }}
                />
            </div>
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'row',
                    justifyContent: 'center',
                    gap: 2,
                }}
            >
                <GrafanaIFrame
                    title='CPU Usage'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsResourcesGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 1,
                    }}
                    externalPathname={podsResourcesGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 1,
                    }}
                />
                <GrafanaIFrame
                    title='Memory Usage'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsResourcesGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 4,
                    }}
                    externalPathname={podsResourcesGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 4,
                    }}
                />
                <GrafanaIFrame
                    title='Receive Bandwidth'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsResourcesGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 6,
                    }}
                    externalPathname={podsResourcesGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 6,
                    }}
                />
                <GrafanaIFrame
                    title='Receive Bandwidth'
                    style={iframeStyle}
                    baseUrl=''
                    pathname={podsResourcesGrafanaPathname}
                    query={{
                        ...baseGrafanaQuery,
                        panelId: 7,
                    }}
                    externalPathname={podsResourcesGrafanaExternalPathname}
                    externalQuery={{
                        ...baseGrafanaQuery,
                        viewPanel: 7,
                    }}
                />
            </div>
        </div>
    )
}
