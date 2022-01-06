/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable import/no-cycle */
import { IBentoSchema, ImageBuildStatus } from './bento'
import { ILabelItemSchema } from './label'
import { IModelRepositorySchema } from './model_repository'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type ModelUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export interface IModelManifestSchema {
    bentoml_version: string
    api_version: string
    module: string
    metadata: {
        [key: string]: any
    }
    context: {
        [key: string]: any
    }
    options: {
        [key: string]: any
    }
    size_bytes: number
}

export interface IModelSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description?: string
    manifest: IModelManifestSchema
    image_build_status: ImageBuildStatus
    upload_status: ModelUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason?: string
    presigned_s3_uri: string
    build_at: string
}

export interface IModelWithRepositorySchema extends IModelSchema {
    repository: IModelRepositorySchema
}

export interface IModelFullSchema extends IModelWithRepositorySchema {
    bentos: IBentoSchema[]
}

export interface ICreateModelSchema {
    version: string
    manifest: IModelManifestSchema
    description?: string
}

export interface IUpdateModelSchema {
    description?: string
    manifest: IModelManifestSchema
    labels: ILabelItemSchema[]
}

export interface IFinishedUploadModelSchema {
    status?: ModelUploadStatus
    reason?: string
}
