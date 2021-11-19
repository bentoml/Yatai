/* eslint-disable @typescript-eslint/no-explicit-any */
/* eslint-disable import/no-cycle */
import { IBentoVersionSchema } from './bento_version'
import { IModelSchema } from './model'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type ModelVersionUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type ModelVersionImageBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IModelVersionManifestSchema {
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
}

export interface IModelVersionSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description?: string
    manifest: IModelVersionManifestSchema
    image_build_status: ModelVersionImageBuildStatus
    upload_status: ModelVersionUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason?: string
    presigned_s3_uri: string
    build_at: string
}

export interface IModelVersionWithModelSchema extends IModelVersionSchema {
    model: IModelSchema
}

export interface IModelVersionFullSchema extends IModelVersionWithModelSchema {
    bento_versions: IBentoVersionSchema[]
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
