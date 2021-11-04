import Card from '@/components/Card'
import KubeResourceDetail from '@/components/KubeResourceDetail'
import Table from '@/components/Table'
import { useFetchYataiComponent } from '@/hooks/useFetchYataiComponent'
import { useFetchYataiComponentHelmChartReleaseResources } from '@/hooks/useFetchYataiComponentHelmChartReleaseResources'
import useTranslation from '@/hooks/useTranslation'
import { IKubeResourceSchema } from '@/schemas/kube_resource'
import { YataiComponentType } from '@/schemas/yatai_component'
import { formatTime } from '@/utils/datetime'
import { useStyletron } from 'baseui'
import { Skeleton } from 'baseui/skeleton'
import { useState } from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import { useParams } from 'react-router-dom'

export default function YataiComponentDetail() {
    const { clusterName, componentType } = useParams<{ clusterName: string; componentType: YataiComponentType }>()

    const [kubeResources, setKubeResources] = useState<IKubeResourceSchema[]>([])
    const [kubeResourcesLoading, setKubeResourcesLoading] = useState(true)

    useFetchYataiComponentHelmChartReleaseResources(
        clusterName,
        componentType,
        setKubeResources,
        setKubeResourcesLoading
    )

    const { yataiComponentInfo } = useFetchYataiComponent(clusterName, componentType)

    const [t] = useTranslation()

    const [, theme] = useStyletron()

    return (
        <div>
            <Card title={t('overview')} titleIcon={RiSurveyLine}>
                <Table
                    isLoading={yataiComponentInfo.isFetching}
                    columns={[
                        t('type'),
                        t('helm release name'),
                        t('helm chart name'),
                        t('helm chart description'),
                        t('created_at'),
                    ]}
                    data={[
                        [
                            yataiComponentInfo.data?.type ?? '',
                            yataiComponentInfo.data?.release?.name ?? '',
                            yataiComponentInfo.data?.release?.chart.metadata.name ?? '',
                            yataiComponentInfo.data?.release?.chart.metadata.description ?? '',
                            yataiComponentInfo.data?.release
                                ? formatTime(yataiComponentInfo.data.release.info.last_deployed)
                                : '-',
                        ],
                    ]}
                />
            </Card>
            <Card
                title={t('kube resources')}
                bodyStyle={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 20,
                }}
            >
                {kubeResourcesLoading ? (
                    <Skeleton rows={3} animation />
                ) : (
                    kubeResources.map((resource, idx) => (
                        <KubeResourceDetail
                            key={idx}
                            resource={resource}
                            clusterName={clusterName}
                            style={{
                                // eslint-disable-next-line @typescript-eslint/no-explicit-any
                                borderBottomStyle: theme.borders.border100.borderStyle as any,
                                borderBottomWidth: theme.borders.border100.borderWidth,
                                borderBottomColor: theme.borders.border200.borderColor,
                            }}
                        />
                    ))
                )}
            </Card>
        </div>
    )
}
