/* eslint-disable import/no-cycle */
import { IBentoRepositorySchema } from './bento_repository'
import { ILabelItemSchema } from './label'
import { IModelSchema } from './model'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type BentoUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type ImageBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IBentoManifestSchema {
    service: string
    bentoml_version: string
    models: string[]
    apis: {
        [key: string]: {
            route: string
            doc: string
            input: string
            output: string
        }
    }
    size_bytes: number
}

export interface IBentoSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description: string
    image_build_status: ImageBuildStatus
    upload_status: BentoUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason: string
    presigned_s3_uri: string
    manifest: IBentoManifestSchema
    build_at: string
}

export interface IBentoWithRepositorySchema extends IBentoSchema {
    repository: IBentoRepositorySchema
}

export interface IBentoFullSchema extends IBentoWithRepositorySchema {
    models: IModelSchema[]
}

export interface ICreateBentoSchema {
    description: string
    version: string
    build_at: string
    manifest: IBentoManifestSchema
}

export interface IUpdateBentoSchema {
    description?: string
    manifest: IBentoManifestSchema
    labels: ILabelItemSchema[]
}

export interface IFinishUploadBentoSchema {
    status?: BentoUploadStatus
    reason?: string
}
