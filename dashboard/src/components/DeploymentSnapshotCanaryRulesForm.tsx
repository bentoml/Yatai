/* eslint-disable jsx-a11y/label-has-associated-control */
import React, { useState } from 'react'
import useTranslation from '@/hooks/useTranslation'
import { DeploymentSnapshotCanaryRuleType, IDeploymentSnapshotCanaryRule } from '@/schemas/deployment_snapshot'
import { Button } from 'baseui/button'
import { DeleteAlt } from 'baseui/icon'
import { Modal, ModalBody, ModalHeader } from 'baseui/modal'
import { Slider } from 'baseui/slider'
import { Input } from 'baseui/input'
import AppDeploymentCanaryRuleTypeSelector from './DeploymentSnapshotCanaryRuleTypeSelector'
import Text from './Text'

interface IDeploymentSnapshotCanaryRulesFormProps {
    value?: IDeploymentSnapshotCanaryRule[]
    onChange?: (value: IDeploymentSnapshotCanaryRule[]) => void
}

export default function DeploymentSnapshotCanaryRulesForm({
    value,
    onChange,
}: IDeploymentSnapshotCanaryRulesFormProps) {
    const [t] = useTranslation()

    const [openModal, setOpenModal] = useState(false)
    const [type, setType] = useState(undefined as DeploymentSnapshotCanaryRuleType | undefined)
    const [weight, setWeight] = useState(undefined as number | undefined)
    const [header, setHeader] = useState(undefined as string | undefined)
    const [cookie, setCookie] = useState(undefined as string | undefined)
    const [headerValue, setHeaderValue] = useState(undefined as string | undefined)

    const formItemStyle = {
        marginTop: 10,
    }

    return (
        <>
            <ul
                style={{
                    margin: '10px 0',
                    listStyle: 'none',
                    paddingLeft: 10,
                }}
            >
                {value?.map((r, idx) => {
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
                        <li
                            key={idx}
                            style={{
                                position: 'relative',
                                display: 'flex',
                                alignItems: 'center',
                                gap: 6,
                                marginBottom: 6,
                            }}
                        >
                            <Button
                                shape='circle'
                                size='mini'
                                onClick={(e) => {
                                    e.preventDefault()
                                    const newValue = (value ?? []).filter((r_) => r_.type !== r.type)
                                    onChange?.(newValue)
                                }}
                            >
                                <DeleteAlt />
                            </Button>
                            <Text>{t(r.type)}:</Text>
                            <Text>{desc}</Text>
                        </li>
                    )
                })}
            </ul>
            <Button
                size='mini'
                overrides={{
                    Root: {
                        style: {
                            marginLeft: 18,
                            marginTop: 10,
                            height: 28,
                        },
                    },
                }}
                onClick={(e) => {
                    e.preventDefault()
                    setOpenModal(true)
                }}
            >
                {t('add')}
            </Button>
            <Modal
                isOpen={openModal}
                onClose={() => {
                    setOpenModal(false)
                }}
            >
                <ModalHeader>{t('add app deployment canary rule')}</ModalHeader>
                <ModalBody>
                    <div style={formItemStyle}>
                        <label>{t('type')}</label>
                        <AppDeploymentCanaryRuleTypeSelector
                            value={type}
                            excludes={value?.map((x) => x.type)}
                            onChange={(type_) => {
                                setType(type_)
                                setWeight(undefined)
                                setHeader(undefined)
                                setHeaderValue(undefined)
                                setCookie(undefined)
                            }}
                        />
                    </div>
                    {type === 'weight' && (
                        <div style={formItemStyle}>
                            <label>{t('weight')}</label>
                            <Slider
                                min={0}
                                max={100}
                                step={1}
                                value={weight !== undefined ? [weight] : [0]}
                                onChange={({ value: value_ }) => {
                                    setWeight(value_[0])
                                }}
                            />
                        </div>
                    )}
                    {type === 'header' && (
                        <div style={formItemStyle}>
                            <label>{t('header')}</label>
                            <Input
                                value={header}
                                onChange={(e) => {
                                    const v = (e.target as HTMLInputElement).value
                                    if (!v) {
                                        return
                                    }
                                    setHeader(v)
                                }}
                            />
                        </div>
                    )}
                    {type === 'header' && (
                        <div style={formItemStyle}>
                            <label>{t('header value')}</label>
                            <Input
                                value={headerValue}
                                onChange={(e) => {
                                    const v = (e.target as HTMLInputElement).value
                                    if (!v) {
                                        return
                                    }
                                    setHeaderValue(v)
                                }}
                            />
                        </div>
                    )}
                    {type === 'cookie' && (
                        <div style={formItemStyle}>
                            <label>{t('cookie')}</label>
                            <Input
                                value={cookie}
                                onChange={(e) => {
                                    const v = (e.target as HTMLInputElement).value
                                    if (!v) {
                                        return
                                    }
                                    setCookie(v)
                                }}
                            />
                        </div>
                    )}
                    <div style={formItemStyle}>
                        <Button
                            size='mini'
                            onClick={(e) => {
                                e.preventDefault()
                                if (!type) {
                                    return
                                }
                                onChange?.([
                                    ...(value || []),
                                    {
                                        type,
                                        weight,
                                        header,
                                        header_value: headerValue,
                                        cookie,
                                    },
                                ])
                                setWeight(undefined)
                                setHeader(undefined)
                                setHeaderValue(undefined)
                                setCookie(undefined)
                                setOpenModal(false)
                            }}
                        >
                            {t('add')}
                        </Button>
                    </div>
                </ModalBody>
            </Modal>
        </>
    )
}
