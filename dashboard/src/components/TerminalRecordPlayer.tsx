import React, { useRef, useEffect } from 'react'

interface ITerminalRecordPlayerProps {
    uid: string
    theme?: string
}

export default function TerminalRecordPlayer({ uid, ...props_ }: ITerminalRecordPlayerProps) {
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    const { asciinema } = window as any

    const { theme } = props_

    let props = props_

    if (!theme) {
        props = {
            ...props,
            theme: 'monokai',
        }
    }
    const ref = useRef(null)

    useEffect(() => {
        const el = ref.current
        if (el) {
            asciinema.player.js.CreatePlayer(el, `/api/v1/terminal_records/${uid}/download`, props)
        }
        return () => {
            if (el) {
                asciinema.player.js.UnmountPlayer(el)
            }
        }
        // eslint-disable-next-line react-hooks/exhaustive-deps
    }, [])

    return <div ref={ref} />
}
