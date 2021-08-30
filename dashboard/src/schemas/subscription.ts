import { ResourceType } from '@/schemas/resource'

export interface ISubscriptionReqSchema {
    action: 'subscribe' | 'unsubscribe'
    resource_type: ResourceType
    resource_uids: string[]
}

export interface ISubscriptionRespSchema<T> {
    resource_type: ResourceType
    payload: T
}
