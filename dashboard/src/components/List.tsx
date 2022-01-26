import React from 'react'
import { Pagination } from 'baseui/pagination'
import { usePage } from '@/hooks/usePage'
import { Skeleton } from 'baseui/skeleton'
import { IPaginationProps } from '@/interfaces/IPaginationProps'
import { useStyletron } from 'baseui'

export interface IListProps<T> {
    isLoading?: boolean
    emptyText?: string
    items: T[]
    onRenderItem: (item: T) => JSX.Element
    itemsContainerClassName?: string
    itemsContainerStyle?: React.CSSProperties
    paginationProps?: IPaginationProps
}

export default function List<T>({
    isLoading = false,
    emptyText,
    items,
    onRenderItem,
    itemsContainerClassName,
    itemsContainerStyle,
    paginationProps,
}: IListProps<T>) {
    const [page, setPage] = usePage()
    const [, theme] = useStyletron()

    return (
        <div>
            <div>
                {isLoading ? (
                    <Skeleton rows={3} height='100px' width='100%' animation />
                ) : (
                    <div className={itemsContainerClassName} style={itemsContainerStyle}>
                        {items.length > 0 ? (
                            items.map((item, idx) => <div key={idx}>{onRenderItem(item)}</div>)
                        ) : (
                            <span
                                style={{
                                    color: theme.colors.contentTertiary,
                                    fontSize: '11px',
                                }}
                            >
                                {emptyText}
                            </span>
                        )}
                    </div>
                )}
            </div>
            {paginationProps && (
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        marginTop: 20,
                    }}
                >
                    <div
                        style={{
                            flexGrow: 1,
                        }}
                    />
                    <Pagination
                        size='mini'
                        numPages={
                            paginationProps.total !== undefined && paginationProps.count !== undefined
                                ? Math.floor(paginationProps.total / paginationProps.count) + 1
                                : 0
                        }
                        currentPage={
                            paginationProps.start !== undefined && paginationProps.count !== undefined
                                ? Math.floor(paginationProps.start / paginationProps.count) + 1
                                : 0
                        }
                        onPageChange={({ nextPage }) => {
                            if (paginationProps.onPageChange) {
                                paginationProps.onPageChange(nextPage)
                            }
                            if (paginationProps.afterPageChange) {
                                setPage({
                                    ...page,
                                    start: (nextPage - 1) * page.count,
                                })
                                paginationProps.afterPageChange(nextPage)
                            }
                        }}
                    />
                </div>
            )}
        </div>
    )
}
