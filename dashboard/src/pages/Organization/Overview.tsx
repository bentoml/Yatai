import React from 'react'
import { RiSurveyLine } from 'react-icons/ri'
import Table from '@/components/Table'
import useTranslation from '@/hooks/useTranslation'
import { useOrganization, useOrganizationLoading } from '@/hooks/useOrganization'
import Card from '@/components/Card'
import { formatDateTime } from '@/utils/datetime'
import User from '@/components/User'

export default function OrganizationOverview() {
    const { organization } = useOrganization()
    const { organizationLoading } = useOrganizationLoading()

    const [t] = useTranslation()

    return (
        <Card title={t('overview')} titleIcon={RiSurveyLine}>
            <Table
                isLoading={organizationLoading}
                columns={[t('name'), t('description'), t('creator'), t('created_at')]}
                data={[
                    [
                        organization?.name,
                        organization?.description,
                        organization?.creator && <User user={organization?.creator} />,
                        organization && formatDateTime(organization.created_at),
                    ],
                ]}
            />
        </Card>
    )
}
