import { useCallback, useMemo } from 'react'
import { IListQuerySchema } from '@/schemas/list'
import { IQueryArgs, IUpdateQueryArgs, useQueryArgs } from './useQueryArgs'

export function usePage(opt?: {
    query?: IQueryArgs
    updateQuery?: IUpdateQueryArgs
    defaultCount?: number
}): [IListQuerySchema, (page: IListQuerySchema) => void] {
    const { query: query_, updateQuery: updateQuery_, defaultCount = 20 } = opt ?? {}
    const { query: query0, updateQuery: updateQuery0 } = useQueryArgs()

    let query = query_
    if (!query) {
        query = query0
    }

    let updateQuery = updateQuery_
    if (!updateQuery) {
        updateQuery = updateQuery0
    }

    const { page: pageStr = '1', search, q, sort_by: sortBy, sort_asc: sortAsc } = query
    let pageNum = parseInt(pageStr, 10)
    // eslint-disable-next-line no-restricted-globals
    if (isNaN(pageNum) || pageNum <= 0) {
        pageNum = 1
    }

    const start = (pageNum - 1) * defaultCount

    return [
        useMemo(
            () => ({
                start,
                count: defaultCount,
                search,
                q,
                sort_by: sortBy,
                sort_asc: sortAsc === 'true',
            }),
            [defaultCount, q, search, sortAsc, sortBy, start]
        ),
        useCallback(
            (newPage) => {
                updateQuery?.({
                    page: Math.floor(newPage.start / newPage.count) + 1,
                    search: newPage.search,
                    q: newPage.q,
                    sort_by: newPage.sort_by,
                    sort_asc: newPage.sort_asc !== undefined ? String(newPage.sort_asc) : undefined,
                })
            },
            [updateQuery]
        ),
    ]
}
