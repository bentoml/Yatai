import useTranslation from '@/hooks/useTranslation'
import { useBentoRepository, useBentoRepositoryLoading } from '@/hooks/useBentoRepository'
import MDEditor from '@uiw/react-md-editor'
import Card from '@/components/Card'
import { MdOutlineDescription } from 'react-icons/md'
import { useCallback, useEffect, useState } from 'react'
import { Skeleton } from 'baseui/skeleton'
import { Button } from 'baseui/button'
import { updateBentoRepository } from '@/services/bento_repository'
import { useParams } from 'react-router-dom'
import BentoListCard from '@/components/BentoListCard'

export default function BentoRepositoryOverview() {
    const { bentoRepositoryName } = useParams<{ bentoRepositoryName: string }>()
    const { bentoRepository, setBentoRepository } = useBentoRepository()
    const { bentoRepositoryLoading } = useBentoRepositoryLoading()
    const [editDescription, setEditDescription] = useState(false)
    const [description, setDescription] = useState(bentoRepository?.description ?? '')
    const [updateLoading, setUpdateLoading] = useState(false)

    const handleUpdateBentoRepository = useCallback(async () => {
        if (!bentoRepository) {
            return
        }
        setUpdateLoading(true)
        try {
            const resp = await updateBentoRepository(bentoRepository.name, {
                ...bentoRepository,
                description,
            })
            setBentoRepository(resp)
            setEditDescription(false)
        } finally {
            setUpdateLoading(false)
        }
    }, [bentoRepository, description, setBentoRepository])

    useEffect(() => {
        if (bentoRepository) {
            setDescription(bentoRepository.description)
        }
    }, [bentoRepository])

    const [t] = useTranslation()

    if (bentoRepositoryLoading) {
        return <Skeleton animation rows={3} />
    }

    return (
        <div>
            <Card
                title={t('description')}
                titleIcon={MdOutlineDescription}
                extra={
                    <div
                        style={{
                            display: 'flex',
                            alignItems: 'center',
                            gap: 10,
                        }}
                    >
                        {editDescription && (
                            <Button
                                kind='secondary'
                                size='compact'
                                onClick={() => {
                                    setEditDescription(false)
                                }}
                            >
                                {t('cancel')}
                            </Button>
                        )}
                        {editDescription && (
                            <Button isLoading={updateLoading} size='compact' onClick={handleUpdateBentoRepository}>
                                {t('submit')}
                            </Button>
                        )}
                        {!editDescription && (
                            <Button
                                size='compact'
                                onClick={() => {
                                    setEditDescription(true)
                                }}
                            >
                                {t('edit')}
                            </Button>
                        )}
                    </div>
                }
            >
                {editDescription ? (
                    <MDEditor value={description} onChange={(v) => setDescription(v ?? '')} />
                ) : (
                    <MDEditor.Markdown source={description} />
                )}
            </Card>
            <BentoListCard bentoRepositoryName={bentoRepositoryName} />
        </div>
    )
}
