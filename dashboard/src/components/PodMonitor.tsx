import { useCluster } from '@/hooks/useCluster'
import { IKubePodSchema } from '@/schemas/kube_pod'
import React from 'react'
import GrafanaIFrame from './GrafanaIFrame'

interface IPodMonitorProps {
    pod?: IKubePodSchema
}

export default function PodMonitor({ pod }: IPodMonitorProps) {
    const iframeStyle = {
        height: 300,
    }

    const { cluster } = useCluster()

    if (!pod || !cluster) {
        return <div>no data</div>
    }

    const podsGrafanaExternalPathname = `${cluster.grafana_root_path}d/resources-pod-2/kubernetes-compute-resources-pod-2`
    const podsGrafanaPathname = `${cluster.grafana_root_path}d-solo/resources-pod-2/kubernetes-compute-resources-pod-2`
    const baseGrafanaQuery = {
        'orgId': 1,
        'from': 'now-12h',
        'to': 'now',
        'var-datasource': 'Prometheus',
        'var-namespace': pod.namespace,
        'var-pod': pod.name,
        'var-container': 'All',
        'fullscreen': true,
        'refresh': '30s',
    }

    return (
        <div>
            <GrafanaIFrame
                title='CPU Usage'
                style={iframeStyle}
                baseUrl=''
                pathname={podsGrafanaPathname}
                query={{
                    ...baseGrafanaQuery,
                    panelId: 1,
                }}
                externalPathname={podsGrafanaExternalPathname}
                externalQuery={{
                    ...baseGrafanaQuery,
                    viewPanel: 1,
                }}
            />
            <GrafanaIFrame
                title='Memory Usage'
                style={iframeStyle}
                baseUrl=''
                pathname={podsGrafanaPathname}
                query={{
                    ...baseGrafanaQuery,
                    panelId: 4,
                }}
                externalPathname={podsGrafanaExternalPathname}
                externalQuery={{
                    ...baseGrafanaQuery,
                    viewPanel: 4,
                }}
            />
            <GrafanaIFrame
                title='Receive Bandwidth'
                style={iframeStyle}
                baseUrl=''
                pathname={podsGrafanaPathname}
                query={{
                    ...baseGrafanaQuery,
                    panelId: 6,
                }}
                externalPathname={podsGrafanaExternalPathname}
                externalQuery={{
                    ...baseGrafanaQuery,
                    viewPanel: 6,
                }}
            />
            <GrafanaIFrame
                title='Receive Bandwidth'
                style={iframeStyle}
                baseUrl=''
                pathname={podsGrafanaPathname}
                query={{
                    ...baseGrafanaQuery,
                    panelId: 7,
                }}
                externalPathname={podsGrafanaExternalPathname}
                externalQuery={{
                    ...baseGrafanaQuery,
                    viewPanel: 7,
                }}
            />
        </div>
    )
}
