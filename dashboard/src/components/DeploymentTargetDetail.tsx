import React from 'react'
import useTranslation from '@/hooks/useTranslation'
import { createUseStyles } from 'react-jss'
import { IDeploymentTargetSchema } from '@/schemas/deployment_target'
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

interface IDeploymentTargetDetailProps {
    deploymentTarget: IDeploymentTargetSchema
}

export default function DeploymentTargetDetail({ deploymentTarget }: IDeploymentTargetDetailProps) {
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
                <span style={valueStyle}>{t(deploymentTarget.type)}</span>
            </div>
            {deploymentTarget.type === 'canary' && (
                <div style={itemStyle}>
                    <Label style={labelStyle}>{t('canary rules')}:</Label>
                    <ul className={styles.rulesWrapper}>
                        {deploymentTarget.canary_rules?.map((r, idx) => {
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
            <div style={itemStyle}>
                <Label style={labelStyle}>{t('replicas')}:</Label>
                <span style={valueStyle}>
                    {deploymentTarget.config?.hpa_conf?.min_replicas} ~{' '}
                    {deploymentTarget.config?.hpa_conf?.max_replicas}
                </span>
            </div>
            <div style={itemStyle}>
                <Label style={labelStyle}>{t('cpu')}:</Label>
                <span style={valueStyle}>
                    {deploymentTarget.config?.resources?.requests?.cpu} ~{' '}
                    {deploymentTarget.config?.resources?.limits?.cpu}
                </span>
            </div>
            <div style={itemStyle}>
                <Label style={labelStyle}>{t('memory')}:</Label>
                <span style={valueStyle}>
                    {deploymentTarget.config?.resources?.requests?.memory} ~{' '}
                    {deploymentTarget.config?.resources?.limits?.memory}
                </span>
            </div>
        </div>
    )
}
