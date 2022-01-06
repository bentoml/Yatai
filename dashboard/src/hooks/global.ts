import { IDeploymentFullSchema } from '@/schemas/deployment'
import { IBentoRepositorySchema } from '@/schemas/bento_repository'
import { IClusterFullSchema } from '@/schemas/cluster'
import { createGlobalState } from 'react-hooks-global-state'
import { IUserSchema } from '@/schemas/user'
import { IOrganizationFullSchema } from '@/schemas/organization'
import { IModelRepositorySchema } from '@/schemas/model_repository'
import { IModelFullSchema } from '@/schemas/model'
import { IBentoFullSchema } from '@/schemas/bento'

export type BaseThemeType = 'light' | 'dark'
export type ThemeType = BaseThemeType | 'followTheSystem'

const initialState = {
    themeType: 'light' as ThemeType,
    currentUser: undefined as IUserSchema | undefined,
    user: undefined as IUserSchema | undefined,
    userLoading: false,
    organization: undefined as IOrganizationFullSchema | undefined,
    organizationLoading: false,
    cluster: undefined as IClusterFullSchema | undefined,
    clusterLoading: false,
    bentoRepository: undefined as IBentoRepositorySchema | undefined,
    bentoRepositoryLoading: false,
    deployment: undefined as IDeploymentFullSchema | undefined,
    deploymentLoading: false,
    modelRepository: undefined as IModelRepositorySchema | undefined,
    modelRepositoryLoading: false,
    model: undefined as IModelFullSchema | undefined,
    modelLoading: false,
    bento: undefined as IBentoFullSchema | undefined,
    bentoLoading: false,
}

const { useGlobalState } = createGlobalState(initialState)
export default useGlobalState
