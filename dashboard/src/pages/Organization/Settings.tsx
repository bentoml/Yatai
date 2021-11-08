import Card from '@/components/Card'
import OrganizationForm from '@/components/OrganizationForm'
import { useOrganization } from '@/hooks/useOrganization'
import useTranslation from '@/hooks/useTranslation'
import { updateOrganization } from '@/services/organization'
import { AiOutlineSetting } from 'react-icons/ai'
import { useCallback } from 'react'
import { Skeleton } from 'baseui/skeleton'

export default function OrganizationSettings() {
    const [t] = useTranslation()
    const { organization, setOrganization } = useOrganization()
    const handleUpdate = useCallback(
        async (values) => {
            if (!organization) {
                return
            }
            const newOrganization = await updateOrganization(values)
            setOrganization(newOrganization)
        },
        [organization, setOrganization]
    )

    return (
        <Card title={t('settings')} titleIcon={AiOutlineSetting}>
            {organization ? (
                <OrganizationForm organization={organization} onSubmit={handleUpdate} />
            ) : (
                <Skeleton animation rows={3} />
            )}
        </Card>
    )
}
