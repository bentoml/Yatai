import { useStyletron } from 'baseui'
import { OnChangeParams, Value } from 'baseui/select'
import React, { useEffect, useRef, useState } from 'react'
import { AiFillCaretDown, AiFillCaretUp } from 'react-icons/ai'
import FilterSelector from './FilterSelector'

export interface IFilterProps {
    label: React.ReactNode
    description?: React.ReactNode
    value?: Value
    options: Value
    onChange?: (params: OnChangeParams) => void
    multiple?: boolean
    showInput?: boolean
}

export default function Filter(props: IFilterProps) {
    const { label, description, value, options, onChange, multiple, showInput } = props
    const [showSelector, setShowSelector] = useState(false)
    const selfClickRef = useRef<boolean>(false)

    useEffect(() => {
        const handleClickOutside = () => {
            if (selfClickRef.current) {
                selfClickRef.current = false
                return
            }
            setShowSelector(false)
        }
        document.addEventListener('click', handleClickOutside)
        return () => {
            document.removeEventListener('click', handleClickOutside)
        }
    }, [])

    const [, theme] = useStyletron()

    return (
        <div
            style={{
                display: 'inline-flex',
                position: 'relative',
            }}
        >
            <div
                role='button'
                tabIndex={0}
                style={{
                    display: 'inline-flex',
                    alignItems: 'center',
                    gap: 4,
                    cursor: 'pointer',
                    padding: '10px 0',
                }}
                onClick={() => {
                    selfClickRef.current = true
                    setShowSelector((v) => !v)
                }}
            >
                <div>{label}</div>
                {showSelector ? <AiFillCaretUp size={12} /> : <AiFillCaretDown size={12} />}
            </div>
            <div
                role='button'
                tabIndex={0}
                style={{
                    position: 'absolute',
                    backgroundColor: theme.colors.backgroundPrimary,
                    top: '100%',
                    bottom: 'auto',
                    left: 'auto',
                    right: 0,
                    zIndex: 999999999999,
                    display: showSelector ? 'block' : 'none',
                    boxShadow: theme.lighting.shadow400,
                    width: 300,
                }}
                onClick={(e) => {
                    e.preventDefault()
                    e.stopPropagation()
                }}
            >
                {description && (
                    <div
                        style={{
                            padding: '10px',
                            fontWeight: 500,
                            borderBottom: `1px solid ${theme.borders.border600.borderColor}`,
                        }}
                    >
                        {description}
                    </div>
                )}
                <FilterSelector
                    showInput={showInput}
                    options={options}
                    value={value}
                    onChange={onChange}
                    multiple={multiple}
                />
            </div>
        </div>
    )
}
