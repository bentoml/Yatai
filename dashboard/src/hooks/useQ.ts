import { Q, parseQ, qToString } from '@/utils'
import { useCallback, useMemo } from 'react'
import { useQueryArgs } from '@/hooks/useQueryArgs'

export function useQ() {
    const { query, updateQuery } = useQueryArgs()
    const qStr = useMemo(() => query.q ?? '', [query])
    const q = useMemo(() => parseQ(qStr), [qStr])

    return {
        q,
        updateQ: useCallback(
            (newQ: Q) => {
                updateQuery({
                    q: qToString({
                        ...q,
                        ...newQ,
                    }),
                })
            },
            [q, updateQuery]
        ),
        replaceQ: useCallback(
            (newQ: Q) => {
                updateQuery({
                    q: qToString(newQ),
                })
            },
            [updateQuery]
        ),
        clearQ: useCallback(() => {
            updateQuery({
                q: undefined,
            })
        }, [updateQuery]),
    }
}
