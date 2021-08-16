import qs from 'qs'
import { useRef } from 'react'
import { useHistory, useLocation } from 'react-router-dom'

export type IQueryArgs = Record<string, string>
export type IUpdateQueryArgs = (query: Record<string, string | number | undefined>) => void

export const useQueryArgs = (): {
    query: IQueryArgs
    updateQuery: IUpdateQueryArgs
} => {
    const location = useLocation()
    const history = useHistory()
    const query = qs.parse(location.search, { ignoreQueryPrefix: true })
    const queryRef = useRef(query)

    const updateQuery = (newQuery: Record<string, string | number | undefined>) => {
        queryRef.current = {
            ...queryRef.current,
            ...Object.keys(newQuery).reduce((p, c) => {
                const v = newQuery[c]
                return {
                    ...p,
                    [c]: v === undefined ? undefined : String(v),
                }
            }, {}),
        }
        history.push({
            pathname: location.pathname,
            search: qs.stringify(queryRef.current, { addQueryPrefix: true }),
        })
    }

    return {
        query: Object.keys(query).reduce((p, c) => {
            const v = query[c]
            if (v === undefined) {
                return p
            }
            return {
                ...p,
                [c]: String(v),
            }
        }, {}),
        updateQuery,
    }
}
