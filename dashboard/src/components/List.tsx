import React from 'react'
import { Pagination } from 'baseui/pagination'
import { usePage } from '@/hooks/usePage'
import { Skeleton } from 'baseui/skeleton'
import { IPaginationProps } from '@/interfaces/IPaginationProps'

export interface IListProps<T> {
    isLoading?: boolean
    items: T[]
    onRenderItem: (item: T) => JSX.Element
    itemsContainerClassName?: string
    itemsContainerStyle?: React.CSSProperties
    paginationProps?: IPaginationProps
}

export default function List<T>({
    isLoading = false,
    items,
    onRenderItem,
    itemsContainerClassName,
    itemsContainerStyle,
    paginationProps,
}: IListProps<T>) {
    const [page, setPage] = usePage()

    return (
        <div>
            <div>
                {isLoading ? (
                    <Skeleton rows={3} height='100px' width='100%' animation />
                ) : (
                    <div className={itemsContainerClassName} style={itemsContainerStyle}>
                        {items.map((item) => onRenderItem(item))}
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
