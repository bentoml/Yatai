import { useCallback, useMemo, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { createBento, listBentos } from '@/services/bento'
import { usePage } from '@/hooks/usePage'
import { ICreateBentoSchema } from '@/schemas/bento'
import BentoForm from '@/components/BentoForm'
import { formatTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import { parseQ, qToString } from '@/utils'
import { useQueryArgs } from '@/hooks/useQueryArgs'
import qs from 'qs'
import FilterBar from './FilterBar'
import FilterInput from './FilterInput'

export default function BentoListCard() {
    const { query, updateQuery } = useQueryArgs()
    const qStr = useMemo(() => query.q ?? '', [query])
    const q = useMemo(() => parseQ(qStr), [qStr])
    const membersInfo = useFetchOrganizationMembers()
    const [page, setPage] = usePage()
    const bentosInfo = useQuery(`fetchClusterBentos:${qs.stringify(page)}`, () => listBentos(page))
    const [isCreateBentoOpen, setIsCreateBentoOpen] = useState(false)
    const handleCreateBento = useCallback(
        async (data: ICreateBentoSchema) => {
            await createBento(data)
            await bentosInfo.refetch()
            setIsCreateBentoOpen(false)
        },
        [bentosInfo]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('bento')])}
            titleIcon={resourceIconMapping.bento}
            middle={
                <div
                    style={{
                        display: 'flex',
                        alignItems: 'center',
                        flexGrow: 1,
                    }}
                >
                    <div
                        style={{
                            width: 100,
                            flexGrow: 1,
                        }}
                    />
                    <div
                        style={{
                            flexGrow: 2,
                            flexShrink: 0,
                            maxWidth: 1200,
                        }}
                    >
                        <FilterInput
                            filterConditions={[
                                {
                                    qStr: 'creator:@me',
                                    label: t('the bentos I created'),
                                },
                                {
                                    qStr: 'last_updater:@me',
                                    label: t('my last updated bentos'),
                                },
                            ]}
                        />
                    </div>
                </div>
            }
            extra={
                <Button size={ButtonSize.compact} onClick={() => setIsCreateBentoOpen(true)}>
                    {t('create')}
                </Button>
            }
        >
            <FilterBar
                filters={[
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],
                        value: ((q.creator as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            q.creator = value.map((v) => String(v.id ?? ''))
                            updateQuery({ q: qToString(q) })
                        },
                        label: t('creator'),
                    },
                    {
                        showInput: true,
                        multiple: true,
                        options:
                            membersInfo.data?.map(({ user }) => ({
                                id: user.name,
                                label: <User user={user} />,
                            })) ?? [],
                        value: ((q.last_updater as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            q.last_updater = value.map((v) => String(v.id ?? ''))
                            updateQuery({ q: qToString(q) })
                        },
                        label: t('last updater'),
                    },
                    {
                        options: [
                            {
                                id: 'updated_at-desc',
                                label: t('newest update'),
                            },
                            {
                                id: 'updated_at-asc',
                                label: t('oldest update'),
                            },
                        ],
                        value: ((q.sort as string[] | undefined) ?? []).map((v) => ({
                            id: v,
                        })),
                        onChange: ({ value }) => {
                            q.sort = value.map((v) => String(v.id ?? ''))
                            updateQuery({ q: qToString(q) })
                        },
                        label: t('sort'),
                    },
                ]}
            />
            <Table
                isLoading={bentosInfo.isLoading}
                columns={[t('name'), t('latest version'), t('last updater'), t('updated_at')]}
                data={
                    bentosInfo.data?.items.map((bento) => [
                        <Link key={bento.uid} to={`/bentos/${bento.name}`}>
                            {bento.name}
                        </Link>,
                        bento.latest_version?.version,
                        bento.latest_version?.creator && <User user={bento.latest_version.creator} />,
                        bento.latest_version?.updated_at && formatTime(bento.latest_version.updated_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentosInfo.data?.start,
                    count: bentosInfo.data?.count,
                    total: bentosInfo.data?.total,
                    onPageChange: ({ nextPage }) => {
                        setPage({
                            ...page,
                            start: nextPage * page.count,
                        })
                        bentosInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateBentoOpen} onClose={() => setIsCreateBentoOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('bento')])}</ModalHeader>
                <ModalBody>
                    <BentoForm onSubmit={handleCreateBento} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
