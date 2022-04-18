import useTranslation from '@/hooks/useTranslation'
import { DeploymentStatus } from '@/schemas/deployment'
import { Spinner } from 'baseui/spinner'
import { Tag, KIND as TagKind, SIZE } from 'baseui/tag'

const statusColorMap: Record<DeploymentStatus, keyof TagKind> = {
    'unknown': TagKind.black,
    'non-deployed': TagKind.primary,
    'running': TagKind.positive,
    'unhealthy': TagKind.warning,
    'failed': TagKind.negative,
    'deploying': TagKind.accent,
    'terminating': TagKind.black,
    'terminated': TagKind.black,
}

export interface IDeploymentStatusProps {
    status: DeploymentStatus
    size?: SIZE[keyof SIZE]
}

export default function DeploymentStatusTag({ status, size = 'small' }: IDeploymentStatusProps) {
    const [t] = useTranslation()
    return (
        <Tag closeable={false} size={size} variant='light' kind={statusColorMap[status]}>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                }}
            >
                {['deploying', 'terminating'].indexOf(status) >= 0 && <Spinner $size={10} />}
                {t(status)}
            </div>
        </Tag>
    )
}
