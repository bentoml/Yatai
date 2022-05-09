import React, { useState } from 'react'
import { Notification } from 'baseui/notification'
import CopyToClipboard from 'react-copy-to-clipboard'
import { Button } from 'baseui/button'
import { TiClipboard } from 'react-icons/ti'
import Text from '@/components/Text'
import useTranslation from '@/hooks/useTranslation'

export interface ICopyableTextProps {
    text: string
}

export default function CopyableText({ text }: ICopyableTextProps) {
    const [copyNotification, setCopyNotification] = useState<string>()

    const [t] = useTranslation()

    return (
        <div
            role='button'
            tabIndex={0}
            style={{
                display: 'flex',
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
                <Text
                    style={{
                        lineHeight: '24px',
                    }}
                >
                    {text}
                </Text>
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
                                    padding: '2px 4px !important',
                                    borderRadius: '2px !important',
                                    fontSize: '11px !important',
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
                    <Button kind='tertiary' size='mini'>
                        <TiClipboard size={12} />
                    </Button>
                </CopyToClipboard>
            </div>
        </div>
    )
}
