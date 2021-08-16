import React from 'react'
import { FieldProps } from 'rc-field-form/lib/Field'
import { Field } from 'rc-field-form'
import styles from './index.module.scss'

export interface IFormItemProps extends FieldProps {
    label?: React.ReactNode
    required?: boolean
}

export const FormItem = ({ label: label_, required, children, ...restProps }: IFormItemProps) => {
    let label = label_
    if (required) {
        label = <span>{label} *</span>
    }
    return (
        <div className={styles.formItem}>
            {/* eslint-disable-next-line react/jsx-props-no-spreading  */}
            <Field {...restProps}>
                {(control, meta, form) => {
                    const childNode =
                        typeof children === 'function'
                            ? children(control, meta, form)
                            : React.cloneElement(children as React.ReactElement, {
                                  label,
                                  errorMessage: meta.errors.join(';'),
                                  ...control,
                              })
                    return <>{childNode}</>
                }}
            </Field>
        </div>
    )
}
