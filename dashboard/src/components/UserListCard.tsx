import Card from '@/components/Card'
import { listUsers } from '@/services/user'
import { usePage } from '@/hooks/usePage'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import User from '@/components/User'
import Table from '@/components/Table'
import { resourceIconMapping } from '@/consts'
import { useQuery } from 'react-query'
import qs from 'qs'

export default function UserListCard() {
    const [page] = usePage()
    const usersInfo = useQuery(`listUsers:${qs.stringify(page)}`, () => {
        return listUsers(page)
    })
    const [t] = useTranslation()

    return (
        <Card title={t('users')} titleIcon={resourceIconMapping.user}>
            <Table
                isLoading={usersInfo.isLoading}
                columns={[t('name'), 'Email', t('created_at')]}
                data={
                    usersInfo.data?.items.map((user) => [
                        <User key={user.uid} user={user} />,
                        user.email,
                        formatDateTime(user.created_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: usersInfo.data?.start,
                    count: usersInfo.data?.count,
                    total: usersInfo.data?.total,
                    afterPageChange: () => {
                        usersInfo.refetch()
                    },
                }}
            />
        </Card>
    )
}
