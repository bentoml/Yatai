import React from 'react'
import { useQuery } from 'react-query'
import { listDeploymentTerminalRecords } from '@/services/deployment'
import { usePage } from '@/hooks/usePage'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { BiPlayCircle } from 'react-icons/bi'
import { Button } from 'baseui/button'
import { StatefulTooltip } from 'baseui/tooltip'
import qs from 'qs'

export interface IDeploymentTerminalRecordListProps {
    clusterName: string
    deploymentName: string
}

export default function DeploymentTerminalRecordList({
    clusterName,
    deploymentName,
}: IDeploymentTerminalRecordListProps) {
    const [page] = usePage()
    const queryKey = `fetchDeploymentTerminalRecords:${clusterName}:${deploymentName}:${qs.stringify(page)}`
    const deploymentTerminalRecordsInfo = useQuery(queryKey, () =>
        listDeploymentTerminalRecords(clusterName, deploymentName, page)
    )

    const [t] = useTranslation()

    return (
        <Table
            isLoading={deploymentTerminalRecordsInfo.isLoading}
            columns={[t('pod'), t('container'), t('creator'), t('created_at'), t('operation')]}
            data={
                deploymentTerminalRecordsInfo.data?.items.map((terminalRecord) => [
                    terminalRecord.pod_name,
                    terminalRecord.container_name,
                    terminalRecord.creator && <User user={terminalRecord.creator} />,
                    formatTime(terminalRecord.created_at),
                    <div key={terminalRecord.uid}>
                        <StatefulTooltip content={t('playback operation')} showArrow>
                            <Button
                                shape='circle'
                                size='mini'
                                onClick={() => {
                                    window.open(
                                        `/clusters/${clusterName}/deployments/${deploymentName}/terminal_records/${terminalRecord.uid}`
                                    )
                                }}
                            >
                                <BiPlayCircle />
                            </Button>
                        </StatefulTooltip>
                    </div>,
                ]) ?? []
            }
            paginationProps={{
                start: deploymentTerminalRecordsInfo.data?.start,
                count: deploymentTerminalRecordsInfo.data?.count,
                total: deploymentTerminalRecordsInfo.data?.total,
                afterPageChange: () => {
                    deploymentTerminalRecordsInfo.refetch()
                },
            }}
        />
    )
}
