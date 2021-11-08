import useTranslation from '@/hooks/useTranslation'
import { DeploymentStatus } from '@/schemas/deployment'
import { StyledSpinnerNext } from 'baseui/spinner'
import { Tag, KIND as TagKind } from 'baseui/tag'

const statusColorMap: Record<DeploymentStatus, keyof TagKind> = {
    'unknown': TagKind.black,
    'non-deployed': TagKind.primary,
    'running': TagKind.positive,
    'unhealthy': TagKind.warning,
    'failed': TagKind.negative,
    'deploying': TagKind.accent,
}

export interface IDeploymentStatusProps {
    status: DeploymentStatus
}

export default function DeploymentStatusTag({ status }: IDeploymentStatusProps) {
    const [t] = useTranslation()
    return (
        <Tag closeable={false} variant='light' kind={statusColorMap[status]}>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    gap: 4,
                }}
            >
                {['deploying'].indexOf(status) >= 0 && <StyledSpinnerNext $size={100} />}
                {t(status)}
            </div>
        </Tag>
    )
}
