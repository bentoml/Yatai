import { YataiComponentType } from '@/schemas/yatai_component'
import { ResourceType } from '@/schemas/resource'
import type { IconBaseProps } from 'react-icons/lib'
import { GrOrganization, GrServerCluster, GrDeploy, GrUser } from 'react-icons/gr'
import { AiOutlineDashboard, AiOutlineCodeSandbox } from 'react-icons/ai'
import { HiOutlineUserGroup, HiOutlineKey } from 'react-icons/hi'
import { BiRevision, BiExtension } from 'react-icons/bi'
import { GoPackage } from 'react-icons/go'
import { VscFileBinary } from 'react-icons/vsc'
import { RiMistFill } from 'react-icons/ri'

export const headerHeight = 55
export const sidebarExpandedWidth = 220
export const sidebarFoldedWidth = 68
export const textVariant = 'smallPlus'
export const dateFormat = 'YYYY-MM-DD'
export const dateWithZeroTimeFormat = 'YYYY-MM-DD 00:00:00'
export const dateTimeFormat = 'YYYY-MM-DD HH:mm:ss'

export const resourceIconMapping: Record<ResourceType, React.ComponentType<IconBaseProps>> = {
    user: GrUser,
    user_group: HiOutlineUserGroup,
    organization: GrOrganization,
    cluster: GrServerCluster,
    bento: GoPackage,
    bento_version: AiOutlineCodeSandbox,
    deployment: GrDeploy,
    deployment_revision: BiRevision,
    yatai_component: BiExtension,
    model: VscFileBinary,
    model_version: VscFileBinary,
    api_token: HiOutlineKey,
}

export const yataiComponentIconMapping: Record<YataiComponentType, React.ComponentType<IconBaseProps>> = {
    logging: RiMistFill,
    monitoring: AiOutlineDashboard,
}
