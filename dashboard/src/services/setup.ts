import axios from 'axios'
import { ISetupSelfHostSchema } from '@/schemas/setup'
import { IUserSchema } from '@/schemas/user'

export async function setupSelfHost(data: ISetupSelfHostSchema): Promise<IUserSchema> {
    const resp = await axios.post('/api/v1/setup', data)
    return resp.data
}
