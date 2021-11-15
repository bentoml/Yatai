// MultipleSelector React Component for displaying a list of items that can be selected

import React, { useMemo, useState } from 'react'
import { OnChangeParams, Value, Option } from 'baseui/select'
import { Input } from 'baseui/input'
import { StatefulMenu } from 'baseui/menu'
import { BsCheck } from 'react-icons/bs'
import _ from 'lodash'
import { useStyletron } from 'baseui'
import { createUseStyles } from 'react-jss'

const useStyles = createUseStyles({
    menu: {
        '& ul li:last-child': {
            borderBottom: 'none',
        },
    },
})

export interface IFilterSelectorProps {
    value?: Value
    options?: Value
    onChange?: (params: OnChangeParams) => void
    multiple?: boolean
    style?: React.CSSProperties
    showInput?: boolean
}

export default function FilterSelector({
    value,
    options,
    onChange,
    multiple = false,
    style,
    showInput,
}: IFilterSelectorProps) {
    const styles = useStyles()
    const [keyword, setKeyword] = useState('')
    const items = useMemo(() => {
        const items_ = (
            options?.map(({ id, label, searchKey }: Option) => {
                return {
                    id,
                    label,
                    searchKey,
                }
            }) ?? []
        ).filter(({ id, label, searchKey: searchKey_ }) => {
            if (!keyword) return true
            let searchKey = ''
            if (typeof searchKey_ === 'string') {
                searchKey = searchKey_
            } else if (typeof label === 'string') {
                searchKey = label
            } else if (typeof id === 'string') {
                searchKey = id
            }
            return searchKey.toLowerCase().includes(keyword.toLowerCase())
        })
        if (showInput) {
            return _.sortBy(items_, (item) => (value?.find((v) => v.id === item.id) ? 0 : 1))
        }
        return items_
    }, [keyword, options, showInput, value])

    const [, theme] = useStyletron()

    return (
        <div
            className={styles.menu}
            style={{
                backgroundColor: theme.colors.backgroundPrimary,
                ...style,
            }}
        >
            <div
                style={{
                    display: showInput ? 'flex' : 'none',
                    padding: 6,
                    borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                }}
            >
                <Input
                    size='mini'
                    value={keyword}
                    clearable
                    onChange={(e) => {
                        setKeyword(e.currentTarget.value)
                    }}
                />
            </div>
            <StatefulMenu
                size='compact'
                items={items}
                overrides={{
                    List: {
                        style: {
                            boxShadow: 'none',
                            maxHeight: '300px',
                            overflowY: 'auto',
                        },
                    },
                    ListItem: {
                        style: {
                            borderBottom: `1px solid ${theme.borders.border200.borderColor}`,
                        },
                    },
                    Option: {
                        props: {
                            getItemLabel: (item: Option) => {
                                const { id, label } = item
                                const contains = value?.find((v) => v.id === id) !== undefined
                                return (
                                    <div
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 10,
                                        }}
                                    >
                                        <div
                                            style={{
                                                display: 'flex',
                                                flexBasis: '20px',
                                                width: '20px',
                                            }}
                                        >
                                            {contains && <BsCheck size={14} />}
                                        </div>
                                        <div>{label}</div>
                                    </div>
                                )
                            },
                        },
                    },
                }}
                onItemSelect={(e) => {
                    let newValue: Value = value ?? []
                    const contains = newValue.find((v) => v.id === e.item.id) !== undefined
                    if (multiple) {
                        if (!contains) {
                            newValue = [...newValue, e.item]
                        } else {
                            newValue = newValue.filter((v) => v.id !== e.item.id)
                        }
                    } else if (!contains) {
                        newValue = [e.item]
                    } else {
                        newValue = []
                    }
                    onChange?.({ value: newValue })
                }}
            />
        </div>
    )
}
