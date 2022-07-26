/* eslint-disable no-nested-ternary */
import Card from '@/components/Card'
import DeploymentMonitor from '@/components/DeploymentMonitor'
import { useDeployment } from '@/hooks/useDeployment'
import useTranslation from '@/hooks/useTranslation'
import { Skeleton } from 'baseui/skeleton'
import { AiOutlineDashboard } from 'react-icons/ai'

export default function DeploymentMonitorPage() {
    const { deployment } = useDeployment()

    const hasMonitoring = false
    const [t] = useTranslation()

    return (
        <Card title={t('monitor')} titleIcon={AiOutlineDashboard}>
            {hasMonitoring ? (
                deployment ? (
                    <DeploymentMonitor deployment={deployment} />
                ) : (
                    <Skeleton animation rows={3} />
                )
            ) : (
                t('please install yatai component first', [t('monitoring')])
            )}
        </Card>
    )
}
