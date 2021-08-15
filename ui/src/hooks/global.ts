import { IClusterSchema } from '@/schemas/cluster'
import { createGlobalState } from 'react-hooks-global-state'
import { IUserSchema } from '@/schemas/user'
import { IOrganizationSchema } from '@/schemas/organization'

export type BaseThemeType = 'light' | 'dark'
export type ThemeType = BaseThemeType | 'followTheSystem'

const initialState = {
    themeType: 'light' as ThemeType,
    currentUser: undefined as IUserSchema | undefined,
    user: undefined as IUserSchema | undefined,
    userLoading: false,
    organization: undefined as IOrganizationSchema | undefined,
    organizationLoading: false,
    cluster: undefined as IClusterSchema | undefined,
    clusterLoading: false,
}

const { useGlobalState } = createGlobalState(initialState)
export default useGlobalState
