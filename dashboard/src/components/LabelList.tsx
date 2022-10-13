import React, { useCallback, useState } from 'react'
import { ILabelItemSchema } from '@/schemas/label'
import { createUseStyles } from 'react-jss'
import { BiEdit } from 'react-icons/bi'
import { Button } from 'baseui/button'
import { RiDeleteBin5Line } from 'react-icons/ri'
import { Input } from 'baseui/input'
import useTranslation from '@/hooks/useTranslation'
import { LabelXSmall } from 'baseui/typography'
import _ from 'lodash'
import { IThemedStyleProps } from '@/interfaces/IThemedStyle'
import { useCurrentThemeType } from '@/hooks/useCurrentThemeType'
import { useStyletron } from 'baseui'
import { IoMdAdd } from 'react-icons/io'

const useStyles = createUseStyles({
    root: {},
    items: {},
    item: (props: IThemedStyleProps) => ({
        'justifyContent': 'space-between',
        'padding': '6px 0',
        'width': '100%',
        'display': 'flex',
        'alignItems': 'center',
        'gap': 10,
        'borderBottom': `1px solid ${props.theme.borders.border100.borderColor}`,
        '&:hover $itemAction': {
            display: 'flex',
        },
        'height': 36,
    }),
    itemContent: {
        display: 'flex',
        alignItems: 'center',
        gap: 10,
    },
    key: {},
    value: {},
    itemAction: {
        display: 'none',
        alignItems: 'center',
        gap: 8,
    },
})

interface ILabelListProps {
    value?: ILabelItemSchema[]
    onChange?: (value: ILabelItemSchema[]) => Promise<void>
    style?: React.CSSProperties
    ignoreKeys?: string[]
}

export default function LabelList({ value = [], ignoreKeys = [], onChange, style }: ILabelListProps) {
    const themeType = useCurrentThemeType()
    const [, theme] = useStyletron()
    const styles = useStyles({ theme, themeType })
    const [editingKey, setEditingKey] = useState<string>()
    const [editingValue, setEditingValue] = useState('')
    const [editLoading, setEditLoading] = useState(false)
    const [deleteLoading, setDeleteLoading] = useState(false)
    const [showAddInputs, setShowAddInputs] = useState(false)
    const [addingKey, setAddingKey] = useState('')
    const [addingValue, setAddingValue] = useState('')
    const [addLoading, setAddLoading] = useState(false)
    const [t] = useTranslation()

    const handleEditInputChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
        setEditingValue(e.target.value)
    }, [])

    const handleAddSubmit = useCallback(async () => {
        if (!addingKey || !addingValue) {
            return
        }
        setAddLoading(true)
        try {
            if (value.find((item) => item.key === addingKey)) {
                onChange?.(value.map((item) => (item.key === addingKey ? { ...item, value: addingValue } : item)))
            } else {
                await onChange?.([
                    ...value,
                    {
                        key: addingKey,
                        value: addingValue,
                    },
                ])
            }
            setAddingKey('')
            setAddingValue('')
            setShowAddInputs(false)
        } finally {
            setAddLoading(false)
        }
    }, [addingKey, addingValue, onChange, value])

    const handleEditSubmit = useCallback(
        async (e) => {
            e.preventDefault()
            setEditLoading(true)
            try {
                await onChange?.(
                    value.map((label) => (label.key === editingKey ? { key: editingKey, value: editingValue } : label))
                )
                setEditingKey(undefined)
            } finally {
                setEditLoading(false)
            }
        },
        [editingKey, editingValue, onChange, value]
    )

    const handleDeleteSubmit = useCallback(
        async (deletingKey: string) => {
            setDeleteLoading(true)
            try {
                await onChange?.(value.filter((label) => label.key !== deletingKey))
            } finally {
                setDeleteLoading(false)
            }
        },
        [onChange, value]
    )

    return (
        <div className={styles.root} style={style}>
            <div className={styles.items}>
                {value.map((label) => (
                    <div
                        className={styles.item}
                        key={label.key}
                        style={{
                            display: ignoreKeys.includes(label.key) ? 'none' : 'flex',
                        }}
                    >
                        <div className={styles.itemContent}>
                            <div className={styles.key}>{label.key}:</div>
                            <div className={styles.value}>
                                {editingKey === label.key ? (
                                    <div
                                        style={{
                                            display: 'flex',
                                            alignItems: 'center',
                                            gap: 12,
                                        }}
                                    >
                                        <Input value={editingValue} onChange={handleEditInputChange} size='mini' />
                                        <div
                                            style={{
                                                display: 'flex',
                                                alignItems: 'center',
                                                gap: 6,
                                                flexShrink: 0,
                                            }}
                                        >
                                            <Button
                                                onClick={(e) => {
                                                    e.preventDefault()
                                                    setEditingKey(undefined)
                                                }}
                                                isLoading={editLoading}
                                                size='mini'
                                                kind='secondary'
                                            >
                                                {t('cancel')}
                                            </Button>
                                            <Button onClick={handleEditSubmit} isLoading={editLoading} size='mini'>
                                                {t('ok')}
                                            </Button>
                                        </div>
                                    </div>
                                ) : (
                                    label.value
                                )}
                            </div>
                        </div>
                        <div className={styles.itemAction}>
                            <Button
                                shape='circle'
                                size='mini'
                                disabled={editingKey === label.key}
                                onClick={(e) => {
                                    e.preventDefault()
                                    setEditingKey(label.key)
                                    setEditingValue(label.value)
                                }}
                            >
                                <BiEdit />
                            </Button>
                            <Button
                                shape='circle'
                                size='mini'
                                isLoading={deleteLoading}
                                onClick={(e) => {
                                    e.preventDefault()
                                    handleDeleteSubmit(label.key)
                                }}
                            >
                                <RiDeleteBin5Line />
                            </Button>
                        </div>
                    </div>
                ))}
            </div>
            <div
                style={{
                    display: 'flex',
                    alignItems: 'center',
                    justifyContent: 'space-between',
                    height: 32,
                    gap: 10,
                    marginTop: 10,
                }}
            >
                <div
                    style={{
                        display: !showAddInputs ? 'none' : 'flex',
                        alignItems: 'center',
                    }}
                >
                    <div
                        style={{
                            display: showAddInputs ? 'flex' : 'none',
                            alignItems: 'center',
                            gap: 10,
                        }}
                    >
                        <div
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 5,
                            }}
                        >
                            <LabelXSmall
                                overrides={{
                                    Block: {
                                        style: {
                                            flexShrink: 0,
                                        },
                                    },
                                }}
                            >
                                {t('key')}
                            </LabelXSmall>
                            <Input
                                size='mini'
                                value={addingKey}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                    setAddingKey(_.trim(e.target.value ?? ''))
                                }
                            />
                        </div>
                        <div
                            style={{
                                display: 'flex',
                                alignItems: 'center',
                                gap: 5,
                            }}
                        >
                            <LabelXSmall
                                overrides={{
                                    Block: {
                                        style: {
                                            flexShrink: 0,
                                        },
                                    },
                                }}
                            >
                                {t('value')}
                            </LabelXSmall>
                            <Input
                                size='mini'
                                value={addingValue}
                                onChange={(e: React.ChangeEvent<HTMLInputElement>) =>
                                    setAddingValue(_.trim(e.target.value ?? ''))
                                }
                            />
                        </div>
                    </div>
                </div>
                <div
                    style={{
                        display: !showAddInputs ? 'none' : 'flex',
                        flexShrink: 0,
                        alignItems: 'center',
                        gap: 10,
                    }}
                >
                    <Button
                        overrides={{
                            Root: {
                                style: {
                                    display: !showAddInputs ? 'none' : 'flex',
                                },
                            },
                        }}
                        isLoading={addLoading}
                        onClick={(e) => {
                            e.preventDefault()
                            setShowAddInputs(false)
                        }}
                        size='mini'
                        kind='secondary'
                    >
                        {t('cancel')}
                    </Button>
                    <Button
                        isLoading={addLoading}
                        overrides={{
                            Root: {
                                style: {
                                    display: !showAddInputs ? 'none' : 'flex',
                                },
                            },
                        }}
                        onClick={(e) => {
                            e.preventDefault()
                            handleAddSubmit()
                        }}
                        size='mini'
                    >
                        {t('ok')}
                    </Button>
                </div>
                <div
                    style={{
                        display: showAddInputs ? 'none' : 'flex',
                        flexShrink: 0,
                        alignItems: 'center',
                        gap: 10,
                    }}
                >
                    <Button
                        isLoading={addLoading}
                        onClick={(e) => {
                            e.preventDefault()
                            setShowAddInputs(true)
                        }}
                        size='mini'
                        startEnhancer={() => <IoMdAdd size={12} />}
                    >
                        {t('add')}
                    </Button>
                </div>
            </div>
        </div>
    )
}
