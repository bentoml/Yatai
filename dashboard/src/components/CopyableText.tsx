import React, { useState } from 'react'
import { Notification } from 'baseui/notification'
import CopyToClipboard from 'react-copy-to-clipboard'
import { Button } from 'baseui/button'
import { TiClipboard } from 'react-icons/ti'
import useTranslation from '@/hooks/useTranslation'
import HighlightText from './HighlightText'

export interface ICopyableTextProps {
    text: string
    highlight?: boolean
}

export default function CopyableText({ text, highlight = false }: ICopyableTextProps) {
    const [copyNotification, setCopyNotification] = useState<string>()

    const [t] = useTranslation()

    return (
        <div
            role='button'
            tabIndex={0}
            style={{
                display: 'inline-flex',
                gap: 3,
            }}
            onClick={(e) => {
                e.stopPropagation()
            }}
        >
            <div
                style={{
                    display: 'flex',
                    flexDirection: 'column',
                    gap: 4,
                }}
            >
                {!highlight ? (
                    <span
                        style={{
                            lineHeight: '16px',
                        }}
                    >
                        {text}
                    </span>
                ) : (
                    <HighlightText>{text}</HighlightText>
                )}
                {copyNotification && (
                    <Notification
                        closeable
                        onClose={() => setCopyNotification(undefined)}
                        kind='positive'
                        overrides={{
                            Body: {
                                style: {
                                    margin: '0 !important',
                                    width: '100%',
                                    boxSizing: 'border-box',
                                    padding: '0px 4px !important',
                                    borderRadius: '2px !important',
                                    fontSize: '11px !important',
                                    lineHeight: '100% !important',
                                    display: 'flex !important',
                                    alignItems: 'center !important',
                                },
                            },
                        }}
                    >
                        {copyNotification}
                    </Notification>
                )}
            </div>
            <div>
                <CopyToClipboard
                    text={text}
                    onCopy={() => {
                        setCopyNotification(t('copied to clipboard'))
                    }}
                >
                    <Button
                        kind='tertiary'
                        size='mini'
                        overrides={{
                            BaseButton: {
                                style: {
                                    'color': 'inherit',
                                    'backgroundColor': 'inherit',
                                    'padding': '0px',
                                    'height': '16px',
                                    ':hover': {
                                        backgroundColor: 'inherit',
                                    },
                                },
                            },
                        }}
                    >
                        <TiClipboard size={12} />
                    </Button>
                </CopyToClipboard>
            </div>
        </div>
    )
}
