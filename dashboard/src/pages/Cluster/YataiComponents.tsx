import Card from '@/components/Card'
import KubeResourcePodStatuses from '@/components/KubeResourcePodStatuses'
import Table from '@/components/Table'
import Time from '@/components/Time'
import YataiComponentTypeRender from '@/components/YataiComponentTypeRender'
import { resourceIconMapping } from '@/consts'
import { useFetchYataiComponents } from '@/hooks/useFetchYataiComponents'
import useTranslation from '@/hooks/useTranslation'
import { YataiComponentType } from '@/schemas/yatai_component'
import { useStyletron } from 'baseui'
import { StatefulTooltip } from 'baseui/tooltip'
import { useParams } from 'react-router-dom'

export default function ClusterYataiComponents() {
    const { clusterName } = useParams<{ clusterName: string }>()
    const { yataiComponentsInfo } = useFetchYataiComponents(clusterName)

    const [t] = useTranslation()
    const [, theme] = useStyletron()

    return (
        <Card title={t('yatai components')} titleIcon={resourceIconMapping.yatai_component}>
            <Table
                isLoading={yataiComponentsInfo.isLoading}
                columns={[t('name'), 'Pods', t('version'), t('installed_at'), t('status')]}
                data={
                    yataiComponentsInfo.data?.map((component) => {
                        let status: 'healthy' | 'unhealthy' = 'unhealthy'
                        if (component.latest_heartbeat_at) {
                            const lastHeartbeat = new Date(component.latest_heartbeat_at).getTime()
                            const now = new Date().getTime()
                            if (now - lastHeartbeat < 60000 * 6) {
                                status = 'healthy'
                            }
                        }
                        return [
                            <YataiComponentTypeRender
                                key={component.name}
                                type={component.name as YataiComponentType}
                            />,
                            <KubeResourcePodStatuses
                                key={component.uid}
                                clusterName={clusterName}
                                resource={{
                                    api_version: 'v1',
                                    kind: 'Deployment',
                                    name: 'deployment',
                                    namespace: component.kube_namespace,
                                    match_labels: component.manifest?.selector_labels ?? {},
                                }}
                            />,
                            component.version,
                            <Time key={component.uid} time={component.latest_installed_at ?? ''} />,
                            status === 'unhealthy' ? (
                                <StatefulTooltip content={t('yatai component unhealthy reason desc')}>
                                    <span
                                        style={{
                                            color: theme.colors.negative,
                                        }}
                                    >
                                        {t(status)}
                                    </span>
                                </StatefulTooltip>
                            ) : (
                                <span
                                    style={{
                                        color: theme.colors.positive,
                                    }}
                                >
                                    {t(status)}
                                </span>
                            ),
                        ]
                    }) ?? []
                }
            />
        </Card>
    )
}
