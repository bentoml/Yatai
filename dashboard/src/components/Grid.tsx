/* eslint-disable react/jsx-props-no-spreading */
import React from 'react'
import { Skeleton } from 'baseui/skeleton'
import { Pagination } from 'baseui/pagination'
import { usePage } from '@/hooks/usePage'
import { useStyletron } from 'baseui'
import { IPaginationProps } from '@/interfaces/IPaginationProps'

export interface IGridProps<T> {
    isLoading?: boolean
    items: T[]
    onRenderItem: (item: T) => JSX.Element
    paginationProps?: IPaginationProps
}

export default function Grid<T>({ isLoading = false, items, onRenderItem, paginationProps }: IGridProps<T>) {
    const [page, setPage] = usePage()

    const [, theme] = useStyletron()

    return (
        <div
            style={{
                paddingTop: 20,
            }}
        >
            <div>
                {isLoading ? (
                    <Skeleton rows={3} animation />
                ) : (
                    <div
                        style={{
                            display: 'grid',
                            gridTemplateColumns: 'repeat(auto-fit, minmax(260px, 400px))',
                            gap: 20,
                        }}
                    >
                        {items.map((item, idx) => (
                            <div
                                key={idx}
                                style={{
                                    position: 'relative',
                                    padding: 10,
                                    border: `1px solid ${theme.borders.border200.borderColor}`,
                                    borderRadius: 2,
                                    boxShadow: theme.lighting.shadow400,
                                }}
                            >
                                {onRenderItem(item)}
                            </div>
                        ))}
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
