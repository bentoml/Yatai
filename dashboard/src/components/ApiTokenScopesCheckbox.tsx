import useTranslation from '@/hooks/useTranslation'
import { ApiTokenScope } from '@/schemas/api_token'
import { useStyletron } from 'baseui'
import { Checkbox } from 'baseui/checkbox'
import React from 'react'

interface IApiTokenScopesCheckboxProps {
    value?: ApiTokenScope[]
    onChange?: (value: ApiTokenScope[]) => void
    style?: React.CSSProperties
}

export default function ApiTokenScopesCheckbox({ value = [], onChange, style }: IApiTokenScopesCheckboxProps) {
    const [t] = useTranslation()
    const [css, theme] = useStyletron()

    return (
        <div style={style}>
            <div>
                <Checkbox
                    onChange={(e) => {
                        const target = e.target as HTMLInputElement
                        if (target.checked) {
                            if (!value?.find((scope) => scope === 'api')) {
                                onChange?.([...(value ?? []), 'api'])
                            }
                        } else {
                            onChange?.(value?.filter((scope) => scope !== 'api') ?? [])
                        }
                    }}
                    checked={!!value.find((scope) => scope === 'api')}
                >
                    {t('api')}
                </Checkbox>
                <div className={css({ padding: theme.sizing.scale400 })}>{t('scope api description')}</div>
            </div>
            <div>
                <Checkbox
                    checked={
                        value.filter((scope) => scope === 'read_organization' || scope === 'write_organization')
                            .length === 2
                    }
                    onChange={(e) => {
                        const target = e.target as HTMLInputElement
                        if (target.checked) {
                            let newValue = value
                            if (!newValue.find((scope) => scope === 'read_organization')) {
                                newValue = [...newValue, 'read_organization']
                            }
                            if (!newValue.find((scope) => scope === 'write_organization')) {
                                newValue = [...newValue, 'write_organization']
                            }
                            onChange?.(newValue)
                        } else {
                            onChange?.(
                                value.filter((scope) => scope !== 'read_organization' && scope !== 'write_organization')
                            )
                        }
                    }}
                    isIndeterminate={
                        value.filter((scope) => scope === 'read_organization' || scope === 'write_organization')
                            .length === 1
                    }
                >
                    {t('organization')}
                </Checkbox>
                <div className={css({ padding: theme.sizing.scale400 })}>
                    <Checkbox
                        checked={!!value.find((scope) => scope === 'read_organization')}
                        onChange={(e) => {
                            const target = e.target as HTMLInputElement
                            if (target.checked) {
                                if (!value.find((scope) => scope === 'read_organization')) {
                                    onChange?.([...value, 'read_organization'])
                                }
                            } else {
                                onChange?.(value.filter((scope) => scope !== 'read_organization'))
                            }
                        }}
                    >
                        {t('read_organization')}
                    </Checkbox>
                    <div className={css({ padding: theme.sizing.scale400 })}>
                        {t('scope read organization description')}
                    </div>
                    <Checkbox
                        checked={!!value.find((scope) => scope === 'write_organization')}
                        onChange={(e) => {
                            const target = e.target as HTMLInputElement
                            if (target.checked) {
                                if (!value.find((scope) => scope === 'write_organization')) {
                                    onChange?.([...value, 'write_organization'])
                                }
                            } else {
                                onChange?.(value.filter((scope) => scope !== 'write_organization'))
                            }
                        }}
                    >
                        {t('write_organization')}
                    </Checkbox>
                    <div className={css({ padding: theme.sizing.scale400 })}>
                        {t('scope write organization description')}
                    </div>
                </div>
            </div>
            <div>
                <Checkbox
                    checked={
                        value.filter((scope) => scope === 'read_cluster' || scope === 'write_cluster').length === 2
                    }
                    onChange={(e) => {
                        const target = e.target as HTMLInputElement
                        if (target.checked) {
                            let newValue = value
                            if (!newValue.find((scope) => scope === 'read_cluster')) {
                                newValue = [...newValue, 'read_cluster']
                            }
                            if (!newValue.find((scope) => scope === 'write_cluster')) {
                                newValue = [...newValue, 'write_cluster']
                            }
                            onChange?.(newValue)
                        } else {
                            onChange?.(value.filter((scope) => scope !== 'read_cluster' && scope !== 'write_cluster'))
                        }
                    }}
                    isIndeterminate={
                        value.filter((scope) => scope === 'read_cluster' || scope === 'write_cluster').length === 1
                    }
                >
                    {t('cluster')}
                </Checkbox>
                <div className={css({ padding: theme.sizing.scale400 })}>
                    <Checkbox
                        checked={!!value.find((scope) => scope === 'read_cluster')}
                        onChange={(e) => {
                            const target = e.target as HTMLInputElement
                            if (target.checked) {
                                if (!value.find((scope) => scope === 'read_cluster')) {
                                    onChange?.([...value, 'read_cluster'])
                                }
                            } else {
                                onChange?.(value.filter((scope) => scope !== 'read_cluster'))
                            }
                        }}
                    >
                        {t('read_cluster')}
                    </Checkbox>
                    <div className={css({ padding: theme.sizing.scale400 })}>{t('scope read cluster description')}</div>
                    <Checkbox
                        checked={!!value.find((scope) => scope === 'write_cluster')}
                        onChange={(e) => {
                            const target = e.target as HTMLInputElement
                            if (target.checked) {
                                if (!value.find((scope) => scope === 'write_cluster')) {
                                    onChange?.([...value, 'write_cluster'])
                                }
                            } else {
                                onChange?.(value.filter((scope) => scope !== 'write_cluster'))
                            }
                        }}
                    >
                        {t('write_cluster')}
                    </Checkbox>
                    <div className={css({ padding: theme.sizing.scale400 })}>
                        {t('scope write cluster description')}
                    </div>
                </div>
            </div>
        </div>
    )
}
