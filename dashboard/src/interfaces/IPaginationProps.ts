export interface IPaginationProps {
    total?: number
    start?: number
    count?: number
    onPageChange?: (page: number) => void
    afterPageChange?: (page: number) => void
}
