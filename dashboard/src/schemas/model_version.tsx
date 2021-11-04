/* eslint-disable import/no-cycle */
import { IModelSchema } from './model'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type ModelVersionUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export interface IModelVersionManifestSchema {
    [key: string]: any // eslint-disable-line @typescript-eslint/no-explicit-any
}

export interface IModelVersionSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description?: string
    manifest: IModelVersionManifestSchema
    uploadStatus: ModelVersionUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason?: string
}

export interface IModelVersionFullSchema extends IModelVersionSchema {
    model: IModelSchema
}

export interface ICreateModelVersionSchema {
    version: string
    manifest: IModelVersionManifestSchema
    description?: string
}

export interface IFinishedUploadModelVersionSchema {
    status?: ModelVersionUploadStatus
    reason?: string
}
