/* eslint-disable react/jsx-props-no-spreading */
import React from 'react'
import { Button, ButtonProps } from 'baseui/button'
import { StatefulTooltip } from 'baseui/tooltip'
import { Block } from 'baseui/block'
import { MdInfoOutline } from 'react-icons/md'

export interface ITooltipButtonProps extends ButtonProps {
    tooltip?: React.ReactNode | ((args: { close: () => void }) => React.ReactNode)
}

export default function TooltipButton({ tooltip, children, ...restProps }: ITooltipButtonProps) {
    return (
        <StatefulTooltip showArrow content={tooltip}>
            <Block
                overrides={{
                    Block: {
                        style: {
                            cursor: 'pointer',
                        },
                    },
                }}
            >
                <Button {...restProps}>
                    <Block
                        overrides={{
                            Block: {
                                style: {
                                    display: 'flex',
                                    alignItems: 'center',
                                    gap: '4px',
                                },
                            },
                        }}
                    >
                        {children}
                        {tooltip ? <MdInfoOutline size={12} /> : undefined}
                    </Block>
                </Button>
            </Block>
        </StatefulTooltip>
    )
}
