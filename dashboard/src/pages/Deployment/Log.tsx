import LokiLog from '@/components/LokiLog'
import React from 'react'

export default function DeploymentLog() {
    return (
        <div
            style={{
                height: 'calc(100vh - 220px)',
            }}
        >
            <LokiLog />
        </div>
    )
}
