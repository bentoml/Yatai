import TerminalRecordPlayer from '@/components/TerminalRecordPlayer'
import React from 'react'
import { useParams } from 'react-router-dom'

export default function DeploymentTerminalRecordPlayer() {
    const { uid } = useParams<{ uid: string }>()
    return <TerminalRecordPlayer uid={uid} />
}
