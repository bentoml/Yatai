import { useCallback, useState } from 'react'
import { useQuery } from 'react-query'
import Card from '@/components/Card'
import { createBentoRepository, listBentoRepositories } from '@/services/bento_repository'
import { usePage } from '@/hooks/usePage'
import { ICreateBentoRepositorySchema } from '@/schemas/bento_repository'
import BentoRepositoryForm from '@/components/BentoRepositoryForm'
import { formatDateTime } from '@/utils/datetime'
import useTranslation from '@/hooks/useTranslation'
import { Button, SIZE as ButtonSize } from 'baseui/button'
import User from '@/components/User'
import { Modal, ModalHeader, ModalBody } from 'baseui/modal'
import Table from '@/components/Table'
import { Link } from 'react-router-dom'
import { resourceIconMapping } from '@/consts'
import { useFetchOrganizationMembers } from '@/hooks/useFetchOrganizationMembers'
import qs from 'qs'
import { useQ } from '@/hooks/useQ'
import FilterBar from './FilterBar'
import FilterInput from './FilterInput'

export default function BentoRepositoryListCard() {
    const { q, updateQ } = useQ()
    const membersInfo = useFetchOrganizationMembers()
    const [page] = usePage()
    const bentoRepositoriesInfo = useQuery(`fetchBentoRepositories:${qs.stringify(page)}`, () =>
        listBentoRepositories(page)
    )
    const [isCreateBentoOpen, setIsCreateBentoOpen] = useState(false)
    const handleCreateBento = useCallback(
        async (data: ICreateBentoRepositorySchema) => {
            await createBentoRepository(data)
            await bentoRepositoriesInfo.refetch()
            setIsCreateBentoOpen(false)
        },
        [bentoRepositoriesInfo]
    )
    const [t] = useTranslation()

    return (
        <Card
            title={t('sth list', [t('bento repository')])}
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
                            updateQ({
                                creator: value.map((v) => String(v.id ?? '')),
                            })
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
                            updateQ({
                                last_updater: value.map((v) => String(v.id ?? '')),
                            })
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
                            updateQ({
                                sort: value.map((v) => String(v.id ?? '')),
                            })
                        },
                        label: t('sort'),
                    },
                ]}
            />
            <Table
                isLoading={bentoRepositoriesInfo.isLoading}
                columns={[t('name'), t('latest version'), t('last updater'), t('updated_at')]}
                data={
                    bentoRepositoriesInfo.data?.items.map((bentoRepository) => [
                        <Link key={bentoRepository.uid} to={`/bento_repositories/${bentoRepository.name}`}>
                            {bentoRepository.name}
                        </Link>,
                        bentoRepository.latest_bento?.version,
                        bentoRepository.latest_bento?.creator && <User user={bentoRepository.latest_bento.creator} />,
                        bentoRepository.latest_bento?.updated_at &&
                            formatDateTime(bentoRepository.latest_bento.updated_at),
                    ]) ?? []
                }
                paginationProps={{
                    start: bentoRepositoriesInfo.data?.start,
                    count: bentoRepositoriesInfo.data?.count,
                    total: bentoRepositoriesInfo.data?.total,
                    afterPageChange: () => {
                        bentoRepositoriesInfo.refetch()
                    },
                }}
            />
            <Modal isOpen={isCreateBentoOpen} onClose={() => setIsCreateBentoOpen(false)} closeable animate autoFocus>
                <ModalHeader>{t('create sth', [t('bento repository')])}</ModalHeader>
                <ModalBody>
                    <BentoRepositoryForm onSubmit={handleCreateBento} />
                </ModalBody>
            </Modal>
        </Card>
    )
}
