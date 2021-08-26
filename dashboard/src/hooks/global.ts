import { IBentoSchema } from '@/schemas/bento'
import { IClusterFullSchema } from '@/schemas/cluster'
import { createGlobalState } from 'react-hooks-global-state'
import { IUserSchema } from '@/schemas/user'
import { IOrganizationFullSchema } from '@/schemas/organization'

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
    bento: undefined as IBentoSchema | undefined,
    bentoLoading: false,
}

const { useGlobalState } = createGlobalState(initialState)
export default useGlobalState
