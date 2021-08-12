import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type BundleVersionUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type BundleVersionBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IBundleVersionSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description: string
    build_status: BundleVersionBuildStatus
    upload_status: BundleVersionUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason: string
    s3_uri: string
}

export interface ICreateBundleVersionSchema {
    description: string
    version: string
}

export interface IFinishUploadBundleVersionSchema {
    status?: BundleVersionUploadStatus
    reason?: string
}
