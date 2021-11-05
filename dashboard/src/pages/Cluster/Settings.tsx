import Card from '@/components/Card'
import ClusterForm from '@/components/ClusterForm'
import { useCluster } from '@/hooks/useCluster'
import useTranslation from '@/hooks/useTranslation'
import { updateCluster } from '@/services/cluster'
import { AiOutlineSetting } from 'react-icons/ai'
import { useCallback } from 'react'
import { Skeleton } from 'baseui/skeleton'

export default function ClusterSettings() {
    const [t] = useTranslation()
    const { cluster, setCluster } = useCluster()
    const handleUpdate = useCallback(
        async (values) => {
            if (!cluster) {
                return
            }
            const newCluster = await updateCluster(cluster.name, values)
            setCluster(newCluster)
        },
        [cluster, setCluster]
    )

    return (
        <Card title={t('settings')} titleIcon={AiOutlineSetting}>
            {cluster ? <ClusterForm cluster={cluster} onSubmit={handleUpdate} /> : <Skeleton animation rows={3} />}
        </Card>
    )
}
