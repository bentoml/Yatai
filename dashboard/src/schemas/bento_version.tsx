import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type BentoVersionUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type BentoVersionBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IBentoVersionSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description: string
    build_status: BentoVersionBuildStatus
    upload_status: BentoVersionUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason: string
    s3_uri: string
}

export interface ICreateBentoVersionSchema {
    description: string
    version: string
}

export interface IFinishUploadBentoVersionSchema {
    status?: BentoVersionUploadStatus
    reason?: string
}
