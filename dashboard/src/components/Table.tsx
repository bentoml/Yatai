import React from 'react'
import { Table as BaseTable, TableProps as BaseTableProps } from 'baseui/table-semantic'
import { Pagination, SIZE as PaginationSize } from 'baseui/pagination'
import { Skeleton } from 'baseui/skeleton'
import { FiInbox } from 'react-icons/fi'
import useTranslation from '@/hooks/useTranslation'
import Text from '@/components/Text'
import { usePage } from '@/hooks/usePage'
import { IPaginationProps } from '@/interfaces/IPaginationProps'

export interface ITableProps extends BaseTableProps {
    paginationProps?: IPaginationProps
}

export default function Table({ isLoading, columns, data, overrides, paginationProps }: ITableProps) {
    const [t] = useTranslation()
    const [page, setPage] = usePage()

    return (
        <>
            <BaseTable
                isLoading={isLoading}
                columns={columns}
                data={data}
                overrides={{
                    TableBodyRow: {
                        style: {
                            cursor: 'pointer',
                        },
                        props: {
                            onClick: (e: React.MouseEvent) => {
                                e.currentTarget.querySelector('a')?.click()
                            },
                        },
                    },
                    ...overrides,
                }}
                loadingMessage={<Skeleton rows={3} height='100px' width='100%' animation />}
                emptyMessage={
                    <div
                        style={{
                            display: 'flex',
                            flexDirection: 'column',
                            alignItems: 'center',
                            justifyContent: 'center',
                            gap: 8,
                        }}
                    >
                        <FiInbox size={30} />
                        <Text>{t('no data')}</Text>
                    </div>
                }
            />
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
                        size={PaginationSize.mini}
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
        </>
    )
}
