/* eslint-disable @typescript-eslint/no-explicit-any */
export interface IHelmChart {
    metadata: {
        name: string
        version: string
        description: string
        icon: string
    }
    values: Record<string, any>
}
