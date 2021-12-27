import { INewsContent } from '@/schemas/news'
import axios from 'axios'

export async function fetchNews(): Promise<INewsContent> {
    const resp = await axios.get<INewsContent>('/api/v1/news')
    return resp.data
}
