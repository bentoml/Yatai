/* eslint-disable import/no-cycle */
import { IBentoRepositorySchema } from './bento_repository'
import { IModelSchema } from './model'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type BentoUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type BentoImageBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IBentoManifestSchema {
    metadata: {
        service_name: string
        service_version: string
        module_name: string
        module_version: string
    }
    apis: {
        name: string
        docs: string
        input_type: string
        output_type: string
    }[]
}

export interface IBentoSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description: string
    image_build_status: BentoImageBuildStatus
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

export interface IFinishUploadBentoSchema {
    status?: BentoUploadStatus
    reason?: string
}
