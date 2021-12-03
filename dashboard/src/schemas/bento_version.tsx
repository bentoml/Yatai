/* eslint-disable import/no-cycle */
import { IBentoSchema } from './bento'
import { IModelVersionSchema } from './model_version'
import { IResourceSchema } from './resource'
import { IUserSchema } from './user'

export type BentoVersionUploadStatus = 'pending' | 'uploading' | 'success' | 'failed'

export type BentoVersionImageBuildStatus = 'pending' | 'building' | 'success' | 'failed'

export interface IBentoVersionManifestSchema {
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

export interface IBentoVersionSchema extends IResourceSchema {
    creator?: IUserSchema
    version: string
    description: string
    image_build_status: BentoVersionImageBuildStatus
    upload_status: BentoVersionUploadStatus
    upload_started_at?: string
    upload_finished_at?: string
    upload_finished_reason: string
    presigned_s3_uri: string
    manifest: IBentoVersionManifestSchema
    build_at: string
}

export interface IBentoVersionWithBentoSchema extends IBentoVersionSchema {
    bento: IBentoSchema
}

export interface IBentoVersionFullSchema extends IBentoVersionWithBentoSchema {
    model_versions: IModelVersionSchema[]
}

export interface ICreateBentoVersionSchema {
    description: string
    version: string
    build_at: string
    manifest: IBentoVersionManifestSchema
}

export interface IFinishUploadBentoVersionSchema {
    status?: BentoVersionUploadStatus
    reason?: string
}
