import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { formatTime } from '@/utils/datetime'
import { createUseStyles } from 'react-jss'
import { IDeploymentSnapshotSchema } from '@/schemas/deployment_snapshot'
import User from './User'
import Label from './Label'

const useStyles = createUseStyles({
    rulesWrapper: {
        'padding': '5px 0',
        'margin': 0,
        'listStyle': 'none',
        '& li': {
            'padding': '6px 0',
            'font-size': '12px',
            'borderBottom': '1px solid #eee',
            '&:last-child': {
                border: 'none',
                paddingBottom: 0,
            },
            '&:first-child': {
                paddingTop: 0,
            },
        },
    },
})

interface IDeploymentSnapshotDetailProps {
    deploymentSnapshot: IDeploymentSnapshotSchema
}

export default function DeploymentSnapshotDetail({ deploymentSnapshot }: IDeploymentSnapshotDetailProps) {
    const labelStyle: React.CSSProperties = {
        width: 100,
        marginRight: 10,
        textAlign: 'right',
    }

    const valueStyle: React.CSSProperties = {
        padding: '5px 0',
        margin: 0,
        listStyle: 'none',
    }

    const itemStyle = {
        marginBottom: 10,
        display: 'flex',
        alignItems: 'center',
    }

    const [t] = useTranslation()

    const styles = useStyles()

    return (
        <div style={{ width: 900 }}>
            <div style={itemStyle}>
                <Label style={labelStyle}>{t('type')}:</Label>
                <span style={valueStyle}>{t(deploymentSnapshot.type)}</span>
            </div>
            {deploymentSnapshot.type === 'canary' && (
                <div style={itemStyle}>
                    <Label style={labelStyle}>{t('canary rules')}:</Label>
                    <ul className={styles.rulesWrapper}>
                        {deploymentSnapshot.canary_rules?.map((r, idx) => {
                            let desc
                            switch (r.type) {
                                case 'weight':
                                    desc = String(r.weight)
                                    break
                                case 'header':
                                    desc = r.header
                                    if (r.header_value) {
                                        desc = `${r.header} = ${r.header_value}`
                                    }
                                    break
                                case 'cookie':
                                    desc = r.cookie
                                    break
                                default:
                                    break
                            }
                            return (
                                <li key={idx}>
                                    {t(r.type)}: {desc}
                                </li>
                            )
                        })}
                    </ul>
                </div>
            )}
            {deploymentSnapshot.creator && (
                <div style={itemStyle}>
                    <Label style={labelStyle}>{t('creator')}:</Label>
                    <User style={valueStyle} user={deploymentSnapshot.creator} size='16px' />
                </div>
            )}
            <div style={itemStyle}>
                <Label style={labelStyle}>{t('created_at')}:</Label>
                <span style={valueStyle}>{formatTime(deploymentSnapshot.created_at)}</span>
            </div>
        </div>
    )
}
