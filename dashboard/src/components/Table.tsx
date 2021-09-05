import React from 'react'
import { Table as BaseTable, TableProps as BaseTableProps } from 'baseui/table-semantic'
import { Pagination, SIZE as PaginationSize, PaginationProps } from 'baseui/pagination'
import { Skeleton } from 'baseui/skeleton'
import { FiInbox } from 'react-icons/fi'
import useTranslation from '@/hooks/useTranslation'
import Text from '@/components/Text'

export interface ITableProps extends BaseTableProps {
    paginationProps?: {
        total?: number
        start?: number
        count?: number
        onPageChange?: PaginationProps['onPageChange']
    }
}

export default function Table({ isLoading, columns, data, overrides, paginationProps }: ITableProps) {
    const [t] = useTranslation()

    return (
        <>
            <BaseTable
                isLoading={isLoading}
                columns={columns}
                data={data}
                overrides={overrides}
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
                        onPageChange={paginationProps.onPageChange}
                    />
                </div>
            )}
        </>
    )
}
